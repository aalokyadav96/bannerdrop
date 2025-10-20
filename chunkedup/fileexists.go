package chunkedup

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/julienschmidt/httprouter"
)

func FileExistsHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	entityType := r.URL.Query().Get("entityType")
	pictureType := r.URL.Query().Get("pictureType")
	entityId := r.URL.Query().Get("entityId")
	fileName := r.URL.Query().Get("fileName")

	if entityType == "" || entityId == "" || fileName == "" {
		http.Error(w, "Missing parameters", http.StatusBadRequest)
		return
	}

	path := filepath.Join("uploads", entityType, entityId, pictureType, fileName)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}

	w.WriteHeader(http.StatusOK)
}
