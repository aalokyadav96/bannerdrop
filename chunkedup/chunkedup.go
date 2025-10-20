package chunkedup

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"naevis/filemgr"

	"github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	uploadDir   = "./uploads"
	tempDir     = "./uploads/tmp"
	maxUpload   = 50 << 20 // 50MB
	cleanupAge  = 2 * time.Minute
	chunkBuffer = 1024 * 256 // 256KB
)

var allowedTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
}

// lockMap ensures concurrency safety per temp folder
var lockMap = struct {
	mu    sync.Mutex
	locks map[string]*sync.Mutex
}{locks: make(map[string]*sync.Mutex)}

type ChunkMeta struct {
	FileName    string              `json:"fileName"`
	ChunkIndex  int                 `json:"chunkIndex"`
	TotalChunks int                 `json:"totalChunks"`
	EntityType  filemgr.EntityType  `json:"entityType"`
	PictureType filemgr.PictureType `json:"pictureType"`
	EntityID    string              `json:"entityId"`
	Token       string              `json:"token"`
}

// type ChunkMeta struct {
// 	FileName    string `json:"fileName"`
// 	ChunkIndex  int    `json:"chunkIndex"`
// 	TotalChunks int    `json:"totalChunks"`
// 	EntityType  string `json:"entityType"`
// 	PictureType string `json:"pictureType"`
// 	EntityID    string `json:"entityId"`
// 	Token       string `json:"token"`
// }

type UploadResponse struct {
	Status   string `json:"status"`
	FileName string `json:"fileName"`
	Chunk    int    `json:"chunk"`
}

func respondWithError(w http.ResponseWriter, status int, msg string) {
	http.Error(w, msg, status)
	fmt.Printf("[%s] %s\n", time.Now().Format(time.RFC3339), msg)
}

// getLock returns a mutex for a specific temp folder
func getLock(folder string) *sync.Mutex {
	lockMap.mu.Lock()
	defer lockMap.mu.Unlock()
	if _, exists := lockMap.locks[folder]; !exists {
		lockMap.locks[folder] = &sync.Mutex{}
	}
	return lockMap.locks[folder]
}

// saveChunk stores a single chunk safely
func saveChunk(file io.Reader, meta ChunkMeta) (string, error) {
	tempFileDir := filepath.Join(tempDir, fmt.Sprintf("%s_%s_%s", meta.FileName, meta.EntityID, meta.Token))
	if err := os.MkdirAll(tempFileDir, os.ModePerm); err != nil {
		return "", err
	}

	lock := getLock(tempFileDir)
	lock.Lock()
	defer lock.Unlock()

	chunkPath := filepath.Join(tempFileDir, fmt.Sprintf("%d.part", meta.ChunkIndex))
	out, err := os.Create(chunkPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	buf := make([]byte, chunkBuffer)
	if _, err := io.CopyBuffer(out, file, buf); err != nil {
		return "", err
	}

	return tempFileDir, nil
}

// mergeChunks merges all chunks safely
func mergeChunks(tempFileDir string, meta ChunkMeta) (string, error) {
	etype := string(meta.EntityType)
	finalDir := filepath.Join(uploadDir, etype)
	if err := os.MkdirAll(finalDir, os.ModePerm); err != nil {
		return "", err
	}

	finalPath := filepath.Join(finalDir, meta.FileName)
	finalFile, err := os.Create(finalPath)
	if err != nil {
		return "", err
	}
	defer finalFile.Close()

	buf := make([]byte, chunkBuffer)

	for i := 0; i < meta.TotalChunks; i++ {
		partPath := filepath.Join(tempFileDir, fmt.Sprintf("%d.part", i))
		partFile, err := os.Open(partPath)
		if err != nil {
			return "", err
		}
		if _, err := io.CopyBuffer(finalFile, partFile, buf); err != nil {
			partFile.Close()
			return "", err
		}
		partFile.Close()
	}

	// Cleanup temp folder
	os.RemoveAll(tempFileDir)

	// Remove lock
	lockMap.mu.Lock()
	delete(lockMap.locks, tempFileDir)
	lockMap.mu.Unlock()

	return finalPath, nil
}

// allChunksUploaded returns true if all chunks exist
func allChunksUploaded(tempFileDir string, totalChunks int) bool {
	for i := 0; i < totalChunks; i++ {
		partPath := filepath.Join(tempFileDir, fmt.Sprintf("%d.part", i))
		if _, err := os.Stat(partPath); err != nil {
			return false
		}
	}
	return true
}

// updateDB updates MongoDB asynchronously
func updateDB(meta ChunkMeta, finalPath string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		updateFields := bson.M{
			"imageUrls":  finalPath,
			"updated_at": time.Now(),
		}
		if err := filemgr.UpdateEntityPicsInDB(ctx, nil, string(meta.EntityType), meta.EntityID, updateFields); err != nil {
			fmt.Printf("[%s] DB update failed: %v\n", time.Now().Format(time.RFC3339), err)
		}
	}()
}

// validateFileType reads first bytes and validates MIME type
func validateFileType(file io.ReadSeeker) error {
	header := make([]byte, 512)
	if _, err := file.Read(header); err != nil {
		return fmt.Errorf("failed to read file header: %v", err)
	}
	file.Seek(0, 0)
	contentType := http.DetectContentType(header)
	if !allowedTypes[contentType] {
		return fmt.Errorf("unsupported file type: %s", contentType)
	}
	return nil
}

func ChunkedUploads(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUpload)

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}

	file, _, err := r.FormFile("chunk")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "chunk not found")
		return
	}
	defer file.Close()

	metaStr := r.FormValue("meta")
	var meta ChunkMeta
	if err := json.Unmarshal([]byte(metaStr), &meta); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid metadata")
		return
	}

	// Only validate the first chunk
	if meta.ChunkIndex == 0 {
		if seeker, ok := file.(io.ReadSeeker); ok {
			if err := validateFileType(seeker); err != nil {
				respondWithError(w, http.StatusBadRequest, err.Error())
				return
			}
		}
	}

	tempFileDir, err := saveChunk(file, meta)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to save chunk")
		return
	}

	var attachments []Attachment

	lock := getLock(tempFileDir)
	lock.Lock()
	if allChunksUploaded(tempFileDir, meta.TotalChunks) {
		finalPath, err := mergeChunks(tempFileDir, meta)
		if err != nil {
			lock.Unlock()
			respondWithError(w, http.StatusInternalServerError, "failed to merge chunks")
			return
		}

		// Open merged file as multipart.File for SaveFileForEntity
		mergedFile, err := os.Open(finalPath)
		if err != nil {
			lock.Unlock()
			respondWithError(w, http.StatusInternalServerError, "failed to open merged file")
			return
		}
		defer mergedFile.Close()

		// Construct a fake FileHeader
		fakeHeader := &multipart.FileHeader{
			Filename: meta.FileName,
			Size:     int64(meta.TotalChunks) * chunkBuffer, // approximate size
		}

		savedName, ext, err := filemgr.SaveFileForEntity(mergedFile, fakeHeader, meta.EntityType, meta.PictureType)
		if err != nil {
			lock.Unlock()
			respondWithError(w, http.StatusInternalServerError, "save failed")
			return
		}

		attachments = append(attachments, Attachment{
			Filename: meta.FileName,
			Path:     savedName + ext,
		})

		// Optionally update DB asynchronously
		updateDB(meta, finalPath)
	}
	lock.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(attachments)
}

type Attachment struct {
	Filename string `bson:"filename" json:"filename"`
	Path     string `bson:"path" json:"path"`
}

// cleanupTempUploads removes old temp folders
func cleanupTempUploads(maxAge time.Duration) {
	files, _ := os.ReadDir(tempDir)
	for _, f := range files {
		info, err := f.Info()
		if err != nil {
			continue
		}
		if time.Since(info.ModTime()) > maxAge {
			os.RemoveAll(filepath.Join(tempDir, f.Name()))
		}
	}
}

// Optional: background cleanup ticker
func StartCleanupTicker() {
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for range ticker.C {
			cleanupTempUploads(cleanupAge)
		}
	}()
}
