package feedproxy

import (
	"fmt"
	"net/http"

	"naevis/filedrop"
	"naevis/filemgr"

	"github.com/julienschmidt/httprouter"
)

// UpdateTweetPost handles the proxied video upload
func UpdateTweetPost(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Call existing media upload handler
	paths, names, resolutions, err := filedrop.HandleMediaUpload(r, "video", filemgr.EntityType("tweet")) // replace EntityType as needed
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to upload media: %v", err), http.StatusInternalServerError)
		return
	}

	// Respond with JSON containing upload info
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status":"success","paths":%q,"names":%q,"resolutions":%v}`, paths, names, resolutions)
}
