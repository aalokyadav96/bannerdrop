package posts

import (
	"naevis/filemgr"
	"naevis/utils"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// image upload endpoint
func UploadImage(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if err := r.ParseMultipartForm(20 << 20); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid form data")
		return
	}
	fileHeader := r.MultipartForm.File["image"]
	if len(fileHeader) == 0 {
		utils.RespondWithError(w, http.StatusBadRequest, "No image uploaded")
		return
	}
	file, err := fileHeader[0].Open()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to open image")
		return
	}
	defer file.Close()

	path, ext, err := filemgr.SaveFileForEntity(file, fileHeader[0], filemgr.EntityPost, filemgr.PicPhoto)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Image save failed")
		return
	}
	_ = ext
	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"url": path})
}
