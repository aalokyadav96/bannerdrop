package filedrop

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"naevis/db"
	"naevis/middleware"
	"naevis/models"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/bson"
)

type Subtitle struct {
	Index   int
	Start   string // Format: "hh:mm:ss.mmm"
	End     string // Format: "hh:mm:ss.mmm"
	Content string
}

func createSubtitleFile(uniqueID string) {
	subtitles := []Subtitle{
		{1, "00:00:00.000", "00:00:01.000", "Welcome to the video!"},
		{2, "00:00:01.001", "00:00:02.000", "In this video, we'll learn how to create subtitles in Go."},
		{3, "00:00:02.001", "00:00:03.000", "Let's get started!"},
	}

	if err := writeVTT(uniqueID, "en", subtitles); err != nil {
		fmt.Printf("subtitle creation failed: %v\n", err)
		return
	}

	fmt.Printf("Subtitle file for %s created.\n", uniqueID)
}

// writeVTT writes a validated subtitle list to a WebVTT file
func writeVTT(uniqueID, lang string, subtitles []Subtitle) error {
	if err := validateSubtitles(subtitles); err != nil {
		return fmt.Errorf("invalid subtitles: %w", err)
	}

	dir := filepath.Join("static", "uploads", "subtitles", uniqueID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}

	filePath := filepath.Join(dir, fmt.Sprintf("%s-%s.vtt", uniqueID, lang))
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("create subtitle file: %w", err)
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	defer w.Flush()

	if _, err := w.WriteString("WEBVTT\n\n"); err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	for _, s := range subtitles {
		line := fmt.Sprintf("%d\n%s --> %s\n%s\n\n", s.Index, s.Start, s.End, s.Content)
		if _, err := w.WriteString(line); err != nil {
			return fmt.Errorf("write subtitle: %w", err)
		}
	}
	return nil
}

// validateSubtitles ensures subtitles are well-formed
func validateSubtitles(subs []Subtitle) error {
	if len(subs) == 0 {
		return errors.New("empty subtitle list")
	}

	for i, s := range subs {
		if s.Index != i+1 {
			return fmt.Errorf("subtitle index out of order at %d (expected %d)", s.Index, i+1)
		}
		start, err := parseTimestamp(s.Start)
		if err != nil {
			return fmt.Errorf("invalid start timestamp at index %d: %w", s.Index, err)
		}
		end, err := parseTimestamp(s.End)
		if err != nil {
			return fmt.Errorf("invalid end timestamp at index %d: %w", s.Index, err)
		}
		if start >= end {
			return fmt.Errorf("start >= end at index %d", s.Index)
		}
		if strings.TrimSpace(s.Content) == "" {
			return fmt.Errorf("empty content at index %d", s.Index)
		}
	}
	return nil
}

// parseTimestamp converts "hh:mm:ss.mmm" to milliseconds
func parseTimestamp(ts string) (int, error) {
	parts := strings.Split(ts, ":")
	if len(parts) != 3 {
		return 0, fmt.Errorf("wrong format")
	}
	h, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}
	m, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, err
	}
	secParts := strings.Split(parts[2], ".")
	if len(secParts) != 2 {
		return 0, fmt.Errorf("invalid seconds part")
	}
	s, err := strconv.Atoi(secParts[0])
	if err != nil {
		return 0, err
	}
	ms, err := strconv.Atoi(secParts[1])
	if err != nil {
		return 0, err
	}

	if m < 0 || m >= 60 || s < 0 || s >= 60 || ms < 0 || ms >= 1000 {
		return 0, fmt.Errorf("invalid ranges")
	}

	total := (((h*60)+m)*60+s)*1000 + ms
	return total, nil
}

// SaveUploadedVTT handles .vtt file upload, parses and normalizes before saving
func SaveUploadedVTT(w http.ResponseWriter, r *http.Request, uniqueID, lang string) (string, error) {
	// Parse multipart form (limit to ~5MB for subtitle files)
	if err := r.ParseMultipartForm(5 << 20); err != nil {
		http.Error(w, "could not parse multipart form", http.StatusBadRequest)
		return "", err
	}

	file, header, err := r.FormFile("subtitle")
	if err != nil {
		http.Error(w, "subtitle file is required", http.StatusBadRequest)
		return "", err
	}
	defer file.Close()

	if !strings.HasSuffix(strings.ToLower(header.Filename), ".vtt") {
		http.Error(w, "only .vtt files are supported", http.StatusBadRequest)
		return "", fmt.Errorf("invalid file type: %s", header.Filename)
	}

	// Save temporary file
	tempPath := filepath.Join(os.TempDir(), header.Filename)
	tempFile, err := os.Create(tempPath)
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	defer tempFile.Close()

	if _, err := io.Copy(tempFile, file); err != nil {
		return "", fmt.Errorf("save temp vtt: %w", err)
	}

	// Parse & validate
	subs, err := parseVTT(tempPath)
	if err != nil {
		return "", fmt.Errorf("parse vtt failed: %w", err)
	}

	// Normalize by rewriting using writeVTT
	if err := writeVTT(uniqueID, lang, subs); err != nil {
		return "", fmt.Errorf("normalize vtt failed: %w", err)
	}

	// Final file path
	finalPath := filepath.Join("static", "uploads", "subtitles", uniqueID, fmt.Sprintf("%s-%s.vtt", uniqueID, lang))
	return finalPath, nil
}

// parseVTT parses a .vtt file into a slice of Subtitle
func parseVTT(filePath string) ([]Subtitle, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read vtt: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	var subs []Subtitle
	var current Subtitle
	state := 0 // 0=expect index, 1=expect timing, 2=expect content

	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			if current.Index != 0 {
				subs = append(subs, current)
				current = Subtitle{}
			}
			state = 0
			continue
		}

		switch state {
		case 0:
			if line == "WEBVTT" {
				continue
			}
			idx, err := strconv.Atoi(line)
			if err != nil {
				return nil, fmt.Errorf("invalid index: %s", line)
			}
			current.Index = idx
			state = 1
		case 1:
			parts := strings.Split(line, " --> ")
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid timing line: %s", line)
			}
			current.Start = parts[0]
			current.End = parts[1]
			state = 2
		case 2:
			if current.Content == "" {
				current.Content = line
			} else {
				current.Content += "\n" + line
			}
		}
	}

	if current.Index != 0 {
		subs = append(subs, current)
	}
	if err := validateSubtitles(subs); err != nil {
		return nil, err
	}
	return subs, nil
}

// UploadSubtitle lets post authors upload a VTT file for their video posts
func UploadSubtitle(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()
	token := r.Header.Get("Authorization")
	claims, err := middleware.ValidateJWT(token)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	postID := ps.ByName("postid")
	lang := ps.ByName("lang") // now comes from URL
	if lang == "" {
		http.Error(w, "language code is required", http.StatusBadRequest)
		return
	}

	// Get the post from DB
	var post models.FeedPost
	if err := db.PostsCollection.FindOne(ctx, bson.M{"postid": postID}).Decode(&post); err != nil {
		http.Error(w, "post not found", http.StatusNotFound)
		return
	}

	// Only author can upload subtitles
	if post.UserID != claims.UserID {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	// Save subtitle file
	path, err := SaveUploadedVTT(w, r, postID, lang)
	if err != nil {
		log.Printf("subtitle upload failed: %v", err)
		http.Error(w, fmt.Sprintf("failed to save subtitle: %v", err), http.StatusInternalServerError)
		return
	}

	// Update DB (nested subtitles map: subtitles.en, subtitles.fr, etc.)
	update := bson.M{"$set": bson.M{fmt.Sprintf("subtitles.%s", lang): path}}
	_, err = db.PostsCollection.UpdateOne(ctx, bson.M{"postid": postID}, update)
	if err != nil {
		http.Error(w, "failed to update subtitles", http.StatusInternalServerError)
		return
	}

	// Respond
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"ok":       true,
		"message":  "Subtitle uploaded successfully",
		"language": lang,
		"path":     path,
	})
}
