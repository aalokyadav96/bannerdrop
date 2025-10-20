package droping

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"naevis/filedrop"
	"naevis/filemgr"
	"naevis/utils"

	"github.com/julienschmidt/httprouter"
)

const maxUploadBytes = 200 << 20 // 200 MB

type Attachment struct {
	Filename    string `bson:"filename" json:"filename"`
	Extn        string `bson:"extn" json:"extn"`
	Key         string `bson:"key" json:"key"`
	Resolutions []int  `bson:"resolutions,omitempty" json:"resolutions,omitempty"`
}

// FiledropHandler handles file uploads via multipart/form-data
func FiledropHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)

	if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/") {
		utils.RespondWithError(w, http.StatusBadRequest, "content-type must be multipart")
		return
	}

	if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid multipart form: "+err.Error())
		return
	}

	if r.MultipartForm == nil || len(r.MultipartForm.File) == 0 {
		utils.RespondWithError(w, http.StatusBadRequest, "no files provided")
		return
	}

	attachments, err := processUploadedFiles(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, attachments)
}

// processUploadedFiles handles all files in the multipart form
func processUploadedFiles(r *http.Request) ([]Attachment, error) {
	var attachments []Attachment
	postType := strings.ToLower(r.FormValue("postType"))

	if postType == "" {
		return nil, fmt.Errorf("missing postType")
	}

	for key, files := range r.MultipartForm.File {
		keyLower := strings.ToLower(key)

		for _, fh := range files {
			att, err := processSingleFile(r, fh, keyLower, postType)
			if err != nil {
				return nil, err
			}
			attachments = append(attachments, att...)
		}
	}

	return attachments, nil
}

// processSingleFile handles individual file based on type
func processSingleFile(r *http.Request, fh *multipart.FileHeader, key, postType string) ([]Attachment, error) {
	if key == "feed" && (postType == "video" || postType == "audio") {
		return handleFeedMediaUpload(r, fh, key)
	}
	return handleRegularUpload(fh, key, postType)
}

// handleFeedMediaUpload handles video/audio feed uploads
func handleFeedMediaUpload(r *http.Request, fh *multipart.FileHeader, key string) ([]Attachment, error) {
	var attachments []Attachment
	src, err := fh.Open()
	if err != nil {
		return nil, fmt.Errorf("cannot open file: %w", err)
	}
	defer src.Close()

	header := make([]byte, 512)
	n, _ := src.Read(header)
	mtype := http.DetectContentType(header[:n])
	_, picType := extensionFromContentType(mtype, key)

	// Save the file
	savedPath, uniqueID, extn, err := filedrop.SaveUploadedFile(fh, filemgr.EntityFeed, picType)
	if err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	uploadDir := filemgr.ResolvePath(filemgr.EntityFeed, picType)

	// Process in background
	go func(savedPath, uploadDir, uniqueID string) {
		res, outPaths, err := filedrop.ProcessVideo(r, savedPath, uploadDir, uniqueID, filemgr.EntityFeed)
		if err != nil {
			log.Printf("[Feed] Video processing failed: %v", err)
			return
		}
		log.Printf("[Feed] Video processed successfully: resolutions=%v, paths=%v", res, outPaths)
	}(savedPath, uploadDir, uniqueID)

	attachments = append(attachments, Attachment{
		Filename: uniqueID,
		Extn:     extn,
		Key:      key,
	})

	return attachments, nil
}

// handleRegularUpload handles images, posters, and audio files
func handleRegularUpload(fh *multipart.FileHeader, key, postType string) ([]Attachment, error) {
	var attachments []Attachment

	// Detect MIME type using file header and extension fallback
	mtype, err := detectFileContentType(fh)
	if err != nil {
		return nil, fmt.Errorf("content type detection failed: %v", err)
	}

	// Reopen file for saving (detectFileContentType already closed it)
	file, err := fh.Open()
	if err != nil {
		return nil, fmt.Errorf("cannot reopen file: %v", err)
	}
	defer file.Close()

	log.Println("postType:", postType)
	log.Println("mtype:", mtype)

	_, picType := extensionFromContentType(mtype, postType)
	log.Println("picType:", picType)

	savedName, ext, err := filemgr.SaveFileForEntity(file, fh, filemgr.EntityType(key), picType)
	if err != nil {
		return nil, fmt.Errorf("filemgr save failed: %v", err)
	}

	attachments = append(attachments, Attachment{
		Filename: savedName,
		Extn:     ext,
		Key:      key,
	})

	return attachments, nil
}

func detectFileContentType(fh *multipart.FileHeader) (string, error) {
	file, err := fh.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	header := make([]byte, 512)
	n, _ := file.Read(header)
	file.Seek(0, io.SeekStart)

	mtype := http.DetectContentType(header[:n])
	ext := strings.ToLower(filepath.Ext(fh.Filename))

	// Get MIME from extension
	extType := mime.TypeByExtension(ext)

	// Use extension-based type if:
	// - DetectContentType misclassifies audio as video
	// - or MIME type is too generic (application/octet-stream)
	if strings.HasPrefix(mtype, "video/") && strings.HasPrefix(extType, "audio/") {
		mtype = extType
	} else if mtype == "application/octet-stream" && extType != "" {
		mtype = extType
	}

	return mtype, nil
}

// extensionFromContentType maps mime types to extensions and PictureType
func extensionFromContentType(ct, postType string) (string, filemgr.PictureType) {
	log.Println("ct", ct, "postType", postType)
	ct = strings.ToLower(ct)
	postType = strings.ToLower(postType)

	switch postType {
	case "poster":
		switch ct {
		case "image/jpeg", "image/jpg":
			return ".jpg", filemgr.PicPoster
		case "image/png":
			return ".png", filemgr.PicPoster
		default:
			return ".jpg", filemgr.PicPoster
		}
	}

	if strings.HasPrefix(ct, "image/") {
		switch ct {
		case "image/jpeg", "image/jpg":
			return ".jpg", filemgr.PicPhoto
		case "image/png":
			return ".png", filemgr.PicPhoto
		default:
			return ".jpg", filemgr.PicPhoto
		}
	}

	if strings.HasPrefix(ct, "video/") {
		switch ct {
		case "video/mp4":
			return ".mp4", filemgr.PicVideo
		case "video/webm":
			return ".webm", filemgr.PicVideo
		default:
			return ".mp4", filemgr.PicVideo
		}
	}

	if postType == "audio" || strings.HasPrefix(ct, "audio/") {
		switch ct {
		case "audio/mp3":
			return ".mp3", filemgr.PicAudio
		case "audio/aac":
			return ".aac", filemgr.PicAudio
		default:
			return ".m4a", filemgr.PicAudio
		}
	}

	return ".bin", filemgr.PicPhoto
}

// writeJSON writes JSON response
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Println("Failed to write JSON:", err)
	}
}

// OptionsHandler handles preflight OPTIONS request
func OptionsHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.WriteHeader(http.StatusNoContent)
}
