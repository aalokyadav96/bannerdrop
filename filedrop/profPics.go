package filedrop

import (
	"context"
	"encoding/json"
	"fmt"
	"naevis/db"
	"naevis/filemgr"
	"naevis/middleware"
	"naevis/rdx"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/bson"
)

// Edit profile picture
func EditProfilePic(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	claims, err := middleware.ValidateJWT(r.Header.Get("Authorization"))
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	pictureUpdates, err := updateAvatars(w, r, claims)
	if err != nil {
		http.Error(w, "Failed to update profile picture", http.StatusInternalServerError)
		return
	}

	if err := ApplyProfileUpdates(claims.UserID, pictureUpdates); err != nil {
		http.Error(w, "Failed to update profile picture", http.StatusInternalServerError)
		return
	}

	InvalidateCachedProfile(claims.Username)

	// Return only the new image name as JSON
	origName, ok := pictureUpdates["avatar"].(string)
	if !ok {
		http.Error(w, "Failed to get updated image name", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"avatar": origName})
}

func updateAvatars(_ http.ResponseWriter, r *http.Request, claims *middleware.Claims) (bson.M, error) {
	update := bson.M{}
	_ = claims
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		return nil, fmt.Errorf("error parsing form data: %w", err)
	}

	file, header, err := r.FormFile("avatar_picture")
	if err != nil {
		return nil, fmt.Errorf("avatar upload failed: %w", err)
	}
	defer file.Close()

	origName, thumbName, err := filemgr.SaveImageWithThumb(file, header, filemgr.EntityUser, filemgr.PicPhoto, 100, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("save image with thumb failed: %w", err)
	}

	update["avatar"] = origName
	update["profile_thumb"] = thumbName

	return update, nil
}

// ApplyProfileUpdates merges multiple bson.M maps into a single update map.
// (Not currently used if you’re only calling UpdateUserByUsername directly.)
func ApplyProfileUpdates(userid string, updates ...bson.M) error {
	finalUpdate := bson.M{}
	for _, u := range updates {
		for k, v := range u {
			finalUpdate[k] = v
		}
	}

	_, err := db.UserCollection.UpdateOne(
		context.TODO(),
		bson.M{"userid": userid},
		bson.M{"$set": finalUpdate},
	)
	return err
}

func InvalidateCachedProfile(username string) error {
	_, err := rdx.RdxDel("profile:" + username)
	return err
}
