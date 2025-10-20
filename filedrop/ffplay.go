package filedrop

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	ffprobeTimeout   = 30 * time.Second
	transcodeTimeout = 10 * time.Minute
	posterTimeout    = 45 * time.Second
	audioTimeout     = 3 * time.Minute
)

// Runner abstracts external command execution for easier testing/mocking.
type Runner interface {
	Run(timeout time.Duration, name string, args ...string) (stdout string, stderr string, err error)
}

type realRunner struct{}

func (realRunner) Run(timeout time.Duration, name string, args ...string) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	var out, errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb

	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		return out.String(), errb.String(), fmt.Errorf("%s timed out after %s", name, timeout)
	}
	return out.String(), errb.String(), err
}

// cmdRunner is the global command runner used by this package.
// Replace in tests to mock ffmpeg/ffprobe results.
var cmdRunner Runner = realRunner{}

// getVideoDimensions returns width and height of a video using ffprobe
func getVideoDimensions(videoPath string) (int, int, error) {
	args := []string{
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height",
		"-of", "csv=p=0",
		videoPath,
	}
	stdout, stderr, err := cmdRunner.Run(ffprobeTimeout, "ffprobe", args...)
	if err != nil {
		return 0, 0, fmt.Errorf("ffprobe getVideoDimensions(%s) failed: %w (stderr=%s)", videoPath, err, stderr)
	}

	parts := strings.Split(strings.TrimSpace(stdout), ",")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("ffprobe getVideoDimensions unexpected output for %s: %q", videoPath, stdout)
	}

	width, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, fmt.Errorf("ffprobe parse width for %s: %w", videoPath, err)
	}

	height, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, fmt.Errorf("ffprobe parse height for %s: %w", videoPath, err)
	}

	return width, height, nil
}

func processVideoResolution(inputPath, outputPath string, targetHeight int) error {
	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("create output dir for %s: %w", outputPath, err)
	}

	// scale=-2:HEIGHT preserves aspect ratio, width auto-adjusted to even number
	scaleFilter := fmt.Sprintf("scale=-2:%d", targetHeight)

	args := []string{
		"-y",
		"-i", inputPath,
		"-vf", scaleFilter,
		"-c:v", "libx264",
		"-crf", "23",
		"-preset", "veryfast",
		"-tune", "zerolatency",
		"-pix_fmt", "yuv420p",
		"-max_muxing_queue_size", "9999",
		"-c:a", "aac",
		"-b:a", "128k",
		"-movflags", "+faststart",
		outputPath,
	}

	stdout, stderr, err := cmdRunner.Run(transcodeTimeout, "ffmpeg", args...)
	if err != nil {
		return fmt.Errorf("ffmpeg transcode %s -> %s (%s) failed: %w (stdout=%s, stderr=%s)",
			inputPath, outputPath, scaleFilter, err, stdout, stderr)
	}
	return nil
}

// CreatePoster extracts a poster JPG for the video.
// Picks a frame at 25% of duration and ensures 16:9 aspect ratio (1280x720) with black padding.
func CreatePoster(videoPath, posterPath string) error {
	if err := os.MkdirAll(filepath.Dir(posterPath), 0o755); err != nil {
		return fmt.Errorf("failed to create poster directory for %s: %w", posterPath, err)
	}

	// Ensure .jpg exactly once
	base := strings.TrimSuffix(posterPath, filepath.Ext(posterPath))
	posterJPG := filepath.ToSlash(base + ".jpg")
	log.Println("CreatePoster:", videoPath, "->", posterJPG)

	duration, err := getVideoDuration(videoPath)
	if err != nil || duration <= 0 {
		// Fallback to a default duration baseline if not available
		log.Printf("CreatePoster duration unavailable for %s: %v", videoPath, err)
		duration = 3.0
	}
	// Pick a stable timestamp at 25% into the video.
	// Clamp to at least 1.0s and at most duration-0.5s
	t := duration * 0.25
	if t < 1.0 {
		t = 1.0
	}
	if t > duration-0.5 {
		t = math.Max(0.0, duration-0.5)
	}
	// Add a tiny random nudge to avoid exact same frame across retries
	t += math.Mod(rand.Float64()*0.2, 0.2) // up to +200ms
	timestamp := formatTimestamp(t)

	// Extract a frame and enforce 16:9 with scaling + black padding
	args := []string{
		"-y",
		"-ss", timestamp,
		"-i", videoPath,
		"-vframes", "1",
		"-q:v", "2",
		"-vf", "scale=w=iw*min(1280/iw\\,720/ih):h=ih*min(1280/iw\\,720/ih),pad=1280:720:(1280-iw*min(1280/iw\\,720/ih))/2:(720-ih*min(1280/iw\\,720/ih))/2:black",
		posterJPG,
	}

	stdout, stderr, err := cmdRunner.Run(posterTimeout, "ffmpeg", args...)
	if err != nil {
		return fmt.Errorf("poster creation failed for %s at %s: %w (stdout=%s, stderr=%s)", videoPath, timestamp, err, stdout, stderr)
	}
	return nil
}

// getVideoDuration returns the video duration in seconds using ffprobe.
func getVideoDuration(path string) (float64, error) {
	args := []string{
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "json",
		path,
	}

	stdout, stderr, err := cmdRunner.Run(ffprobeTimeout, "ffprobe", args...)
	if err != nil {
		return 0, fmt.Errorf("ffprobe getVideoDuration(%s) failed: %w (stderr=%s)", path, err, stderr)
	}

	var result struct {
		Format struct {
			Duration string `json:"duration"`
		} `json:"format"`
	}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		return 0, fmt.Errorf("ffprobe unmarshal duration for %s: %w (stdout=%s)", path, err, stdout)
	}
	if strings.TrimSpace(result.Format.Duration) == "" {
		return 0, fmt.Errorf("ffprobe duration not found for %s (stdout=%s)", path, stdout)
	}

	dur, err := strconv.ParseFloat(result.Format.Duration, 64)
	if err != nil {
		return 0, fmt.Errorf("parse duration for %s: %w (value=%s)", path, err, result.Format.Duration)
	}
	return dur, nil
}

// formatTimestamp converts seconds (e.g. 12.345) to "hh:mm:ss.SSS"
func formatTimestamp(seconds float64) string {
	if seconds < 0 {
		seconds = 0
	}
	totalMs := int(seconds * 1000.0)
	h := totalMs / 3600000
	m := (totalMs % 3600000) / 60000
	s := (totalMs % 60000) / 1000
	ms := totalMs % 1000
	return fmt.Sprintf("%02d:%02d:%02d.%03d", h, m, s, ms)
}

func ExtractVideoDuration(videoPath string) float64 {
	res, err := getVideoDuration(videoPath)
	if err != nil {
		return 0.0
	}
	return res
}

// processAudioResolutions converts the input to normalized MP3 and returns the chosen bitrate (kbps)
// and the output path. It probes source bitrate to avoid upscaling if the input is lower.
// It also applies EBU R128 loudness normalization (loudnorm).
func processAudioResolutions(originalFilePath, uploadDir, uniqueID string) ([]int, string) {
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		fmt.Printf("audio: failed to create output dir %s: %v\n", uploadDir, err)
		return []int{}, originalFilePath
	}
	outputPath := filepath.Join(uploadDir, uniqueID+".mp3")

	// Probe input audio bitrate (in bits/s)
	inputBitrate := probeAudioBitrate(originalFilePath)

	// Choose target bitrate: cap at 128 kbps, but don't upscale if input is lower and known
	targetKbps := 128
	if inputBitrate > 0 {
		inKbps := inputBitrate / 1000
		if inKbps < targetKbps {
			targetKbps = inKbps
		}
		if targetKbps <= 0 {
			targetKbps = 128 // fallback
		}
	}

	args := []string{
		"-y",
		"-i", originalFilePath,
		"-vn",                // no video
		"-c:a", "libmp3lame", // convert to mp3
		"-b:a", fmt.Sprintf("%dk", targetKbps),
		"-filter:a", "loudnorm", // loudness normalization
		outputPath,
	}

	stdout, stderr, err := cmdRunner.Run(audioTimeout, "ffmpeg", args...)
	if err != nil {
		fmt.Printf("audio processing failed for %s -> %s: %v\nstdout: %s\nstderr: %s\n", originalFilePath, outputPath, err, stdout, stderr)
		return []int{}, originalFilePath
	}

	return []int{targetKbps}, outputPath
}

// probeAudioBitrate returns the audio stream bitrate in bits/s via ffprobe; 0 if unknown.
func probeAudioBitrate(path string) int {
	args := []string{
		"-v", "error",
		"-select_streams", "a:0",
		"-show_entries", "stream=bit_rate",
		"-of", "json",
		path,
	}
	stdout, _, err := cmdRunner.Run(ffprobeTimeout, "ffprobe", args...)
	if err != nil {
		return 0
	}

	var result struct {
		Streams []struct {
			BitRate json.Number `json:"bit_rate"`
		} `json:"streams"`
	}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil || len(result.Streams) == 0 {
		return 0
	}

	brStr := string(result.Streams[0].BitRate)
	brStr = strings.TrimSpace(brStr)
	if brStr == "" {
		return 0
	}
	br, err := strconv.Atoi(brStr)
	if err != nil || br <= 0 {
		return 0
	}
	return br
}
