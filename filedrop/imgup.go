package filedrop

import (
	"fmt"
	"mime/multipart"
	"naevis/filemgr"
	"naevis/models"
	"naevis/mq"
	"net/http"
)

// -------------------- Multi-image Upload --------------------

func saveUploadedFiles(r *http.Request, formKey, fileType string, entitytype filemgr.EntityType) ([]string, error) {
	files := r.MultipartForm.File[formKey]
	if len(files) == 0 {
		return nil, fmt.Errorf("no %s files uploaded", fileType)
	}

	var ids []string
	entity := entitytype
	picType := filemgr.PictureType(fileType)

	for _, file := range files {
		origName, err := processSingleImageUpload(file, entity, picType)
		if err != nil {
			return nil, fmt.Errorf("failed to process %s: %w", fileType, err)
		}
		ids = append(ids, origName)
	}

	mq.Notify("postpics-uploaded", models.Index{})
	mq.Notify("thumbnail-created", models.Index{})

	return ids, nil
}

func processSingleImageUpload(file *multipart.FileHeader, entity filemgr.EntityType, picType filemgr.PictureType) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("cannot open image: %w", err)
	}
	defer src.Close()

	origName, ext, err := filemgr.SaveFileForEntity(src, file, entity, picType)
	if err != nil {
		return "", fmt.Errorf("saving image failed: %w", err)
	}
	return origName + ext, nil
}
