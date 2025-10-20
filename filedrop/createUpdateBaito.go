package filedrop

import (
	"fmt"
	"log"
	"naevis/db"
	"naevis/filemgr"
	"naevis/utils"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type entityMeta struct {
	Collection *mongo.Collection
	IDField    string
	Prefix     filemgr.EntityType
	OwnerField string
}

// Global metadata map to resolve entity configurations dynamically
var entityMetaMap = map[string]entityMeta{
	"place":  {db.PlacesCollection, "placeid", filemgr.EntityPlace, "createdBy"},
	"event":  {db.EventsCollection, "eventid", filemgr.EntityEvent, "creatorid"},
	"baito":  {db.BaitoCollection, "baitoid", filemgr.EntityBaito, "ownerId"},
	"worker": {db.BaitoWorkerCollection, "baito_user_id", filemgr.EntityWorker, "userid"},
	"artist": {db.ArtistsCollection, "artistid", filemgr.EntityArtist, "creatorid"},
	"farm":   {db.FarmsCollection, "farmid", filemgr.EntityFarm, "createdBy"},
	"crop":   {db.CropsCollection, "cropid", filemgr.EntityCrop, "createdby"},
	"user":   {db.UserCollection, "userid", filemgr.EntityUser, "userid"},
	"recipe": {db.RecipeCollection, "recipeid", filemgr.EntityRecipe, "userId"},
}

// parseImagesForm handles both keepImages and new file uploads
func parseImagesForm(r *http.Request, existingImages []string, entityPrefix filemgr.EntityType) ([]string, error) {
	if err := r.ParseMultipartForm(20 << 20); err != nil {
		return nil, err
	}
	defer r.MultipartForm.RemoveAll()

	// Save new files
	newImages, _ := filemgr.SaveFormFiles(r.MultipartForm, "images", entityPrefix, filemgr.PicPhoto, false)

	// Parse kept images
	keepImages := r.MultipartForm.Value["keepImages"]
	finalImages := []string{}

	if len(keepImages) > 0 {
		finalImages = append(finalImages, keepImages...)
	} else {
		finalImages = existingImages
	}

	// Merge new + kept
	if len(newImages) > 0 {
		finalImages = append(finalImages, newImages...)
	}

	return finalImages, nil
}

// UpdateGalleryImages handles updating images for any entity
func UpdateGalleryImages(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()
	entityType := ps.ByName("entityType")
	entityID := ps.ByName("entityId")
	userID := utils.GetUserIDFromRequest(r)

	meta, ok := entityMetaMap[entityType]
	if !ok {
		utils.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("Unsupported entity type: %s", entityType))
		return
	}

	// Fetch existing document
	var existing struct {
		Images []string `bson:"images"`
	}
	filter := bson.M{
		meta.IDField:    entityID,
		meta.OwnerField: userID,
	}

	err := meta.Collection.FindOne(ctx, filter).Decode(&existing)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, fmt.Sprintf("%s not found or unauthorized", entityType))
		return
	}

	// Parse uploaded & kept images
	finalImages, err := parseImagesForm(r, existing.Images, meta.Prefix)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid form data")
		return
	}

	// Update only the images field
	update := bson.M{
		"$set": bson.M{
			"images":    finalImages,
			"updatedAt": time.Now(),
		},
	}

	_, err = meta.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Printf("[%s] Image update error: %v", entityType, err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update images")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message":    fmt.Sprintf("%s images updated successfully", entityType),
		"entityType": entityType,
		"entityId":   entityID,
	})
}
