package filedrop

import (
	"fmt"
	"mime/multipart"
	"naevis/filemgr"
	"net/http"
	"path/filepath"
	"strings"
)

type MediaType string

const (
	Video MediaType = "video"
	Audio MediaType = "audio"
)

// -------------------- Unified Media Result --------------------

type MediaResult struct {
	Resolutions []int
	Paths       []string
	IDs         []string
}

// -------------------- Processors --------------------

type mediaProcessor func(r *http.Request, savedPath, uploadDir, uniqueID string, entity filemgr.EntityType) ([]int, []string, error)

var mediaPicTypes = map[MediaType]filemgr.PictureType{
	Video: filemgr.PicVideo,
	Audio: filemgr.PicAudio,
}

var mediaProcessors = map[MediaType]mediaProcessor{
	Video: ProcessVideo,
	Audio: func(r *http.Request, savedPath, uploadDir, uniqueID string, entity filemgr.EntityType) ([]int, []string, error) {
		res, paths := processAudio(savedPath, uploadDir, uniqueID, entity)
		return res, paths, nil
	},
}

// -------------------- Media Upload --------------------

func ProcessMediaUpload(r *http.Request, formKey string, mediaType MediaType, entity filemgr.EntityType) (*MediaResult, error) {
	file, err := getUploadedFile(r, formKey)
	if err != nil || file == nil {
		return nil, fmt.Errorf("no file uploaded: %w", err)
	}

	picType, ok := mediaPicTypes[mediaType]
	if !ok {
		return nil, fmt.Errorf("unsupported media type: %s", mediaType)
	}

	savedPath, uniqueID, _, err := SaveUploadedFile(file, entity, picType)
	if err != nil {
		return nil, err
	}

	processor, ok := mediaProcessors[mediaType]
	if !ok {
		return nil, fmt.Errorf("no processor for media type: %s", mediaType)
	}

	res, paths, err := processor(r, savedPath, filemgr.ResolvePath(entity, picType), uniqueID, entity)
	if err != nil {
		return nil, err
	}
	return &MediaResult{
		Resolutions: res,
		Paths:       paths,
		IDs:         []string{uniqueID},
	}, nil
}

// -------------------- File Helpers --------------------

func getUploadedFile(r *http.Request, formKey string) (*multipart.FileHeader, error) {
	if r.MultipartForm == nil {
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			return nil, fmt.Errorf("failed to parse form: %w", err)
		}
	}
	files := r.MultipartForm.File[formKey]
	if len(files) == 0 {
		return nil, nil
	}
	return files[0], nil
}

func SaveUploadedFile(file *multipart.FileHeader, entity filemgr.EntityType, picType filemgr.PictureType) (string, string, string, error) {
	src, err := file.Open()
	if err != nil {
		return "", "", "", fmt.Errorf("cannot open uploaded file: %w", err)
	}
	defer src.Close()

	savedName, ext, err := filemgr.SaveFileForEntity(src, file, entity, picType)
	if err != nil {
		return "", "", "", fmt.Errorf("file save failed: %w", err)
	}

	savedPath := filepath.Join(filemgr.ResolvePath(entity, picType), savedName+ext)
	uniqueID := strings.TrimSuffix(savedName, ext)
	return savedPath, uniqueID, ext, nil
}

func normalizePath(p string) string {
	if !strings.HasPrefix(p, "/") {
		p = "/" + filepath.ToSlash(p)
	}
	return p
}

func generateFilePath(baseDir, uniqueID, extension string) string {
	return filepath.Join(baseDir, fmt.Sprintf("%s.%s", uniqueID, extension))
}
