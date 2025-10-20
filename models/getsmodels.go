package models

import (
	"time"
)

type PlacesResponse struct {
	PlaceID        string   `json:"placeid"`
	Name           string   `json:"name"`
	ShortDesc      string   `json:"short_desc"`
	Address        string   `json:"address,omitempty"`
	Distance       float64  `json:"distance,omitempty"`
	OperatingHours []string `json:"operatinghours,omitempty"`
	Category       string   `json:"category"`
	Tags           []string `json:"tags"`
	Banner         string   `json:"banner"`
}

/* ---------- MODELS ---------- */

type BaitosResponse struct {
	BaitoId      string    `bson:"baitoid,omitempty" json:"baitoid"`
	Title        string    `bson:"title" json:"title"`
	Description  string    `bson:"description" json:"description"`
	Category     string    `bson:"category" json:"category"`
	SubCategory  string    `bson:"subcategory" json:"subcategory"`
	Location     string    `bson:"location" json:"location"`
	Wage         string    `bson:"wage" json:"wage"`
	Requirements string    `bson:"requirements" json:"requirements"`
	BannerURL    string    `bson:"banner,omitempty" json:"banner,omitempty"`
	WorkHours    string    `bson:"workHours" json:"workHours"`
	CreatedAt    time.Time `bson:"createdAt" json:"createdAt"`
	OwnerID      string    `bson:"ownerId" json:"ownerId"`
}

type BaitoWorkersResponse struct {
	UserID      string    `json:"userid" bson:"userid"`
	BaitoUserID string    `json:"baito_user_id" bson:"baito_user_id"`
	Name        string    `json:"name" bson:"name"`
	Age         int       `json:"age" bson:"age"`
	Phone       string    `json:"phone_number" bson:"phone_number"`
	Location    string    `json:"address" bson:"address"`
	Preferred   []string  `json:"preferred_roles" bson:"preferred_roles"`
	Bio         string    `json:"bio" bson:"bio"`
	ProfilePic  string    `json:"photo" bson:"photo"`
	CreatedAt   time.Time `json:"created_at" bson:"created_at"`
}

// --- BlogPostResponse for list view ---
type BlogPostResponse struct {
	PostID      string    `bson:"postid" json:"postid"`
	Title       string    `bson:"title" json:"title"`
	Category    string    `bson:"category" json:"category"`
	Subcategory string    `bson:"subcategory" json:"subcategory"`
	ReferenceID *string   `bson:"referenceId,omitempty" json:"referenceId,omitempty"`
	Thumb       string    `bson:"thumb" json:"thumb"`
	CreatedBy   string    `bson:"createdBy" json:"createdBy"`
	CreatedAt   time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time `bson:"updatedAt" json:"updatedAt"`
}
