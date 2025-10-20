package filedrop

import (
	"fmt"
	"io"
	"naevis/filemgr"
	"naevis/models"
	"naevis/mq"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// -------------------- Video Processing --------------------
func ProcessVideo(r *http.Request, savedPath, uploadDir, uniqueID string, entitytype filemgr.EntityType) ([]int, []string, error) {
	width, height, err := getVideoDimensions(savedPath)
	if err != nil {
		_ = os.Remove(savedPath)
		return nil, nil, fmt.Errorf("failed to get video dimensions: %w", err)
	}

	resolutions, outputPaths := processVideoResolutionsParallel(savedPath, uploadDir, uniqueID, width, height, 3)
	if len(outputPaths) == 0 {
		_ = os.Remove(savedPath)
		return nil, nil, fmt.Errorf("video transcoding failed")
	}

	// posterDir now points directly to poster root, no subfolder per uniqueID
	posterDir := filemgr.ResolvePath(entitytype, filemgr.PicPoster)
	if err := os.MkdirAll(posterDir, 0755); err != nil {
		for _, out := range outputPaths {
			_ = os.Remove(strings.TrimPrefix(filepath.FromSlash(out), string(filepath.Separator)))
		}
		_ = os.Remove(savedPath)
		return nil, nil, fmt.Errorf("failed to create poster directory: %w", err)
	}

	thumbPath := filepath.Join(posterDir, uniqueID+".jpg")

	// Check if user uploaded a thumbnail
	thumbnailFile, _, thumbErr := r.FormFile("thumbnail")
	if thumbErr == nil {
		defer thumbnailFile.Close()

		// Save temporary original upload
		tmpThumb := filepath.Join(os.TempDir(), uniqueID+"_raw_thumb")
		tmpFile, err := os.Create(tmpThumb)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create temp thumbnail: %w", err)
		}
		if _, err := io.Copy(tmpFile, thumbnailFile); err != nil {
			tmpFile.Close()
			_ = os.Remove(tmpThumb)
			return nil, nil, fmt.Errorf("failed to write temp thumbnail: %w", err)
		}
		tmpFile.Close()

		// Normalize to 16:9 with black padding
		args := []string{
			"-y",
			"-i", tmpThumb,
			"-vf", "scale=w=iw*min(1280/iw\\,720/ih):h=ih*min(1280/iw\\,720/ih),pad=1280:720:(1280-iw*min(1280/iw\\,720/ih))/2:(720-ih*min(1280/iw\\,720/ih))/2:black",
			thumbPath,
		}
		stdout, stderr, err := cmdRunner.Run(time.Minute, "ffmpeg", args...)
		_ = os.Remove(tmpThumb)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to process thumbnail: %w (stdout=%s, stderr=%s)", err, stdout, stderr)
		}
	} else {
		// No thumbnail provided → create poster from video
		if err := CreatePoster(savedPath, filepath.Join(posterDir, uniqueID+".jpg")); err != nil {
			for _, out := range outputPaths {
				_ = os.Remove(strings.TrimPrefix(filepath.FromSlash(out), string(filepath.Separator)))
			}
			_ = os.Remove(savedPath)
			return nil, nil, fmt.Errorf("poster creation failed: %w", err)
		}
	}

	go createSubtitleFile(uniqueID)
	mq.Notify("postpics-uploaded", models.Index{})

	return resolutions, outputPaths, nil
}

// -------------------- Video Resolutions --------------------

func processVideoResolutionsParallel(originalFilePath, uploadDir, uniqueID string, origWidth, origHeight int, maxParallel int) ([]int, []string) {
	if maxParallel <= 0 {
		maxParallel = 2
	}
	_ = origWidth

	ladder := []struct {
		Label  string
		Height int
	}{
		{"4320", 4320}, {"2160", 2160}, {"1440", 1440},
		{"1080", 1080}, {"720", 720}, {"480", 480},
		{"360", 360}, {"240", 240}, {"144", 144},
	}

	type task struct {
		Label      string
		Height     int
		OutputPath string
	}
	var tasks []task
	for _, r := range ladder {
		if r.Height > origHeight {
			continue // skip higher than source
		}
		out := generateFilePath(uploadDir, uniqueID+"-"+r.Label, "mp4")
		tasks = append(tasks, task{
			Label:      r.Label,
			Height:     r.Height,
			OutputPath: out,
		})
	}

	if len(tasks) == 0 {
		return nil, nil
	}

	type result struct {
		ok        bool
		height    int
		outputURL string
	}
	taskCh := make(chan task)
	resCh := make(chan result)
	doneCh := make(chan struct{})

	workers := maxParallel
	if workers > len(tasks) {
		workers = len(tasks)
	}
	for i := 0; i < workers; i++ {
		go func() {
			for t := range taskCh {
				err := processVideoResolution(originalFilePath, t.OutputPath, t.Height)
				if err != nil {
					fmt.Printf("Skipping %s due to error: %v\n", t.Label, err)
					resCh <- result{ok: false}
					continue
				}
				urlPath := "/" + filepath.ToSlash(t.OutputPath)
				resCh <- result{ok: true, height: t.Height, outputURL: urlPath}
			}
		}()
	}

	go func() {
		for _, t := range tasks {
			taskCh <- t
		}
		close(taskCh)
	}()

	var heights []int
	var outputs []string
	count := 0
	for res := range resCh {
		count++
		if res.ok {
			heights = append(heights, res.height)
			outputs = append(outputs, res.outputURL)
		}
		if count == len(tasks) {
			close(doneCh)
			break
		}
	}
	<-doneCh
	close(resCh)

	type pair struct {
		h int
		p string
	}
	var pairs []pair
	for i := range heights {
		pairs = append(pairs, pair{h: heights[i], p: outputs[i]})
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[i].h > pairs[j].h })
	heights = heights[:0]
	outputs = outputs[:0]
	for _, pr := range pairs {
		heights = append(heights, pr.h)
		outputs = append(outputs, pr.p)
	}

	return heights, outputs
}
