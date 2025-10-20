package filedrop

import (
	"log"
	"naevis/filemgr"
	"net/http"
)

func HandleMediaUpload(r *http.Request, postType string, entitytype filemgr.EntityType) (paths, names []string, resolutions []int, err error) {
	switch postType {
	case "image":
		names, err = saveUploadedFiles(r, "images", "photo", entitytype)
	case "video":
		var result *MediaResult
		result, err = saveUploadedVideoFile(r, "video", entitytype)
		log.Println("res", result, err)
		if err == nil {
			resolutions, paths, names = result.Resolutions, result.Paths, result.IDs
		}
	case "audio":
		var result *MediaResult
		result, err = saveUploadedAudioFile(r, "audio", entitytype)
		if err == nil {
			resolutions, paths, names = result.Resolutions, result.Paths, result.IDs
		}
	}
	return
}

func saveUploadedVideoFile(r *http.Request, formKey string, entitytype filemgr.EntityType) (*MediaResult, error) {
	return ProcessMediaUpload(r, formKey, Video, entitytype)
}

func saveUploadedAudioFile(r *http.Request, formKey string, entitytype filemgr.EntityType) (*MediaResult, error) {
	return ProcessMediaUpload(r, formKey, Audio, entitytype)
}

// func HandleMediaUpload(r *http.Request, postType string, entitytype filemgr.EntityType) (paths, names []string, resolutions []int, err error) {
// 	switch postType {
// 	case "image":
// 		names, err = saveUploadedFiles(r, "images", "photo", entitytype)
// 	case "video":
// 		resolutions, paths, names, err = saveUploadedVideoFile(r, "video", entitytype)
// 	case "audio":
// 		resolutions, paths, names, err = saveUploadedAudioFile(r, "audio", entitytype)
// 	}
// 	return
// }
