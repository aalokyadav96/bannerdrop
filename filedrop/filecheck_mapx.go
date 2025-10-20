package filedrop

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"mime/multipart"
	"naevis/db"
	"naevis/middleware"
	"naevis/models"
	"net/http"
	"os"
	"slices"

	"github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/bson"
)

// type FileMetadata struct {
// 	ID        primitive.ObjectID  `bson:"_id,omitempty"`
// 	Hash      string              `bson:"hash"`
// 	UserPosts map[string][]string `bson:"userPosts"` // Maps userID to an array of postIDs
// 	PostURLs  map[string]string   `bson:"postUrls"`  // Maps postID to its corresponding URL
// }

func UploadFile(file multipart.File, filePath, userID, postID string) {
	// Compute SHA-256 hash
	hash := ComputeFileHash(file)

	// Update or create a document with the given hash
	updateResult, err := db.FilesCollection.UpdateOne(
		context.TODO(),
		bson.M{"hash": hash}, // Find by hash
		bson.M{
			"$addToSet": bson.M{
				"userPosts." + userID: postID, // Append the postID to the user's posts array
			},
			"$set": bson.M{
				"postUrls." + postID: filePath, // Add or update the postID -> URL mapping
			},
		},
	)
	if err != nil {
		// Handle error during update
		return
	}

	// If no document was updated, create a new one
	if updateResult.MatchedCount == 0 {
		newFile := models.FileMetadata{
			Hash:      hash,
			UserPosts: map[string][]string{userID: {postID}},
			PostURLs:  map[string]string{postID: filePath},
		}
		_, err = db.FilesCollection.InsertOne(context.TODO(), newFile)
		if err != nil {
			// Handle error during insertion
			return
		}
	}
}

func ComputeFileHash(file multipart.File) string {
	hasher := sha256.New()
	file.Seek(0, io.SeekStart) // Reset file pointer before reading
	io.Copy(hasher, file)
	return hex.EncodeToString(hasher.Sum(nil))
}

func RemoveUserFile(userID, postID, hash string) {
	// Pull the postID from the user's posts array
	result, err := db.FilesCollection.UpdateOne(
		context.TODO(),
		bson.M{"hash": hash},
		bson.M{"$pull": bson.M{"userPosts." + userID: postID}}, // Remove the specific postID
	)

	if err != nil || result.MatchedCount == 0 {
		// Handle error or no matching document
		return
	}

	// Check if the postID is no longer associated with any user
	var file models.FileMetadata
	err = db.FilesCollection.FindOne(context.TODO(), bson.M{"hash": hash}).Decode(&file)
	if err == nil {
		isPostAssociated := false
		for _, posts := range file.UserPosts {
			if slices.Contains(posts, postID) {
				isPostAssociated = true
				break
			}
		}

		// Remove the postID and its URL if no association remains
		if !isPostAssociated {
			_, _ = db.FilesCollection.UpdateOne(
				context.TODO(),
				bson.M{"hash": hash},
				bson.M{"$unset": bson.M{"postUrls." + postID: ""}}, // Remove the URL mapping
			)
		}
	}

	// If no users are left, delete the file and its storage
	if len(file.UserPosts) == 0 {
		os.Remove(file.PostURLs[postID]) // Remove the file from disk
		db.FilesCollection.DeleteOne(context.TODO(), bson.M{"hash": hash})
	}
}

func CheckUserInFile(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	tokenString := r.Header.Get("Authorization")
	claims, err := middleware.ValidateJWT(tokenString)
	if err != nil {
		log.Printf("JWT validation error: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Hash string `json:"hash"`
	}

	req.Hash = ps.ByName("hash")

	// // Single decoding attempt
	// if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
	// 	log.Printf("Request decode error: %v", err)
	// 	http.Error(w, "Invalid request", http.StatusBadRequest)
	// 	return
	// }

	// log.Printf("Decoded request: %+v", req)

	checkUserInFile(w, claims.UserID, req.Hash)
}

func checkUserInFile(w http.ResponseWriter, userID, hash string) {
	// Query MongoDB for the file using the hash

	var file models.FileMetadata
	err := db.FilesCollection.FindOne(context.TODO(), bson.M{"hash": hash}).Decode(&file)
	if err != nil {
		// If file is not found or any error occurs
		json.NewEncoder(w).Encode(map[string]any{"exists": false})
		return
	}

	// Check if the userID exists in userPosts
	posts, exists := file.UserPosts[userID]
	if exists && len(posts) > 0 {
		// // Fetch URLs for the user's posts
		// urls := make(map[string]string)
		// for _, postID := range posts {
		// 	if url, ok := file.PostURLs[postID]; ok {
		// 		urls[postID] = url
		// 	}
		// }

		// Return the metadata
		json.NewEncoder(w).Encode(map[string]any{
			"exists": true,
			// "postUrls": urls, // PostID-URL mapping for the user
		})
		return
	}

	// If userID is not found or has no associated posts
	json.NewEncoder(w).Encode(map[string]any{"exists": false})
}
