package droping

import (
	"encoding/json"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
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
	postType := strings.ToLower(strings.TrimSpace(r.FormValue("postType"))) // optional, normalized

	for key, files := range r.MultipartForm.File {
		keyLower := strings.ToLower(key)

		for _, fh := range files {
			atts, err := processSingleFile(r, fh, keyLower, postType)
			if err != nil {
				return nil, err
			}
			attachments = append(attachments, atts...)
		}
	}

	return attachments, nil
}

// processSingleFile handles individual file based on type
func processSingleFile(r *http.Request, fh *multipart.FileHeader, key, postType string) ([]Attachment, error) {
	// When key == "feed" we want special handling:
	// - if postType is empty, default to "video"
	// - if postType == "poster" treat as regular upload
	// - if postType == "video" or "audio" use feed media upload
	if key == "feed" {
		if postType == "" {
			postType = "video"
			log.Printf("[Filedrop] postType missing for feed, defaulting to %q", postType)
		}
		pt := strings.ToLower(strings.TrimSpace(postType))
		if pt == "poster" {
			return handleRegularUpload(fh, key, pt)
		}
		if pt == "video" || pt == "audio" {
			return handleFeedMediaUpload(r, fh, key, pt)
		}
		// fallthrough to regular handling for unexpected postTypes
		return handleRegularUpload(fh, key, pt)
	}

	// For all other keys, ignore postType entirely (but pass it along)
	return handleRegularUpload(fh, key, postType)
}

// handleFeedMediaUpload handles video/audio feed uploads
func handleFeedMediaUpload(r *http.Request, fh *multipart.FileHeader, key, postType string) ([]Attachment, error) {
	var attachments []Attachment

	src, err := fh.Open()
	if err != nil {
		return nil, fmt.Errorf("cannot open file: %w", err)
	}
	// we don't need to keep src open after SaveUploadedFile (it will reopen internally if needed)
	_ = src.Close()

	_, picType := extensionFromContentType(postType)

	// Save the file
	savedPath, uniqueID, extn, err := filedrop.SaveUploadedFile(fh, filemgr.EntityFeed, picType)
	if err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	uploadDir := filemgr.ResolvePath(filemgr.EntityFeed, picType)

	// Process in background (non-blocking). It's fine to capture savedPath/uploadDir/uniqueID only.
	go func(savedPath, uploadDir, uniqueID string) {
		res, outPaths, err := filedrop.ProcessVideo(r, savedPath, uploadDir, uniqueID, filemgr.EntityFeed)
		if err != nil {
			log.Printf("[Feed] Video processing failed for %s: %v", uniqueID, err)
			return
		}
		log.Printf("[Feed] Video processed successfully for %s: resolutions=%v, paths=%v", uniqueID, res, outPaths)
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

	// Reopen file for saving
	file, err := fh.Open()
	if err != nil {
		return nil, fmt.Errorf("cannot reopen file: %v", err)
	}
	defer file.Close()

	log.Println("postType:", postType)

	_, picType := extensionFromContentType(postType)
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

// func detectFileContentType(fh *multipart.FileHeader) (string, error) {
// 	file, err := fh.Open()
// 	if err != nil {
// 		return "", err
// 	}
// 	defer file.Close()

// 	header := make([]byte, 512)
// 	n, _ := file.Read(header)
// 	if seeker, ok := file.(io.Seeker); ok {
// 		_, _ = seeker.Seek(0, io.SeekStart)
// 	}

// 	mtype := http.DetectContentType(header[:n])
// 	ext := strings.ToLower(filepath.Ext(fh.Filename))

// 	// Get MIME from extension
// 	extType := mime.TypeByExtension(ext)

// 	// Use extension-based type if:
// 	// - DetectContentType misclassifies audio as video
// 	// - or MIME type is too generic (application/octet-stream)
// 	if strings.HasPrefix(mtype, "video/") && strings.HasPrefix(extType, "audio/") {
// 		mtype = extType
// 	} else if mtype == "application/octet-stream" && extType != "" {
// 		mtype = extType
// 	}

// 	return mtype, nil
// }

// extensionFromContentType maps mime types / postType strings to extensions and PictureType
func extensionFromContentType(postType string) (string, filemgr.PictureType) {
	log.Println("postType", postType)
	postType = strings.ToLower(strings.TrimSpace(postType))

	switch postType {
	case "audio":
		return "", filemgr.PicAudio
	case "video":
		return "", filemgr.PicVideo
	case "poster":
		return "", filemgr.PicPoster
	}
	return "", filemgr.PicPhoto
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
