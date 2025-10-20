package filedrop

import (
	"encoding/json"
	"log"
	"net/http"

	"naevis/filemgr"
	"naevis/utils"

	"github.com/julienschmidt/httprouter"
)

func UploadHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// 1) Validate JWT
	_ = utils.GetUserIDFromRequest(r)

	// 2) Parse multipart form
	if err := r.ParseMultipartForm(32 << 20); err != nil { // allow up to 32MB total
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// room := strings.TrimSpace(r.FormValue("chat"))
	// if room == "" {
	// 	http.Error(w, "room missing", http.StatusBadRequest)
	// 	return
	// }

	hdrs := r.MultipartForm.File["file"]
	if len(hdrs) == 0 {
		http.Error(w, "no files provided", http.StatusBadRequest)
		return
	}

	// 3) Save each file
	var attachments []Attachment
	for _, hdr := range hdrs {
		file, err := hdr.Open()
		if err != nil {
			http.Error(w, "file error", http.StatusBadRequest)
			return
		}
		savedName, ext, err := filemgr.SaveFileForEntity(file, hdr, filemgr.EntityChat, filemgr.PicPhoto)
		file.Close()
		if err != nil {
			log.Println("save failed:", err)
			http.Error(w, "save failed", http.StatusInternalServerError)
			return
		}
		attachments = append(attachments, Attachment{
			Filename: hdr.Filename,
			Path:     savedName + ext,
		})
	}

	data, _ := json.Marshal(attachments)
	// 6) Respond to HTTP client
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

type Attachment struct {
	Filename string `bson:"filename" json:"filename"`
	Path     string `bson:"path" json:"path"`
}
