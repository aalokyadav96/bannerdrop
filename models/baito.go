package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Baito struct {
	BaitoId          string    `bson:"baitoid,omitempty" json:"baitoid"`
	EntityType       string    `bson:"entityType" json:"entityType"`
	EntityID         string    `bson:"entityId" json:"entityId"`
	Title            string    `bson:"title" json:"title"`
	Description      string    `bson:"description" json:"description"`
	Category         string    `bson:"category" json:"category"`
	SubCategory      string    `bson:"subcategory" json:"subcategory"`
	Location         string    `bson:"location" json:"location"`
	Wage             string    `bson:"wage" json:"wage"`
	Phone            string    `bson:"phone" json:"phone"`
	Requirements     string    `bson:"requirements" json:"requirements"`
	BannerURL        string    `bson:"banner,omitempty" json:"banner,omitempty"`
	Images           []string  `bson:"images" json:"images"`
	WorkHours        string    `bson:"workHours" json:"workHours"`
	Benefits         string    `bson:"benefits,omitempty" json:"benefits,omitempty"`
	Email            string    `bson:"email,omitempty" json:"email,omitempty"`
	Tags             []string  `bson:"tags,omitempty" json:"tags,omitempty"`
	LastDateToApply  time.Time `bson:"lastdate" json:"lastdate"`
	CreatedAt        time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt        time.Time `bson:"updatedAt,omitempty" json:"updatedAt,omitempty"`
	OwnerID          string    `bson:"ownerId" json:"ownerId"`
	ApplicationCount int       `bson:"applicationcount" json:"applicationcount"`
}

type BaitoApplication struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	BaitoID     string             `bson:"baitoid" json:"baitoid"`
	UserID      string             `bson:"userid" json:"userid"`
	Username    string             `bson:"username" json:"username"`
	Pitch       string             `bson:"pitch" json:"pitch"`
	SubmittedAt time.Time          `bson:"submittedAt" json:"submittedAt"`
}

type BaitoWorker struct {
	UserID       string    `json:"userid" bson:"userid"`
	BaitoUserID  string    `json:"baito_user_id" bson:"baito_user_id"`
	Name         string    `json:"name" bson:"name"`
	Age          int       `json:"age" bson:"age"`
	Phone        string    `json:"phone_number" bson:"phone_number"`
	Location     string    `json:"address" bson:"address"`
	Preferred    []string  `json:"preferred_roles" bson:"preferred_roles"`
	Bio          string    `json:"bio" bson:"bio"`
	ProfilePic   string    `json:"photo" bson:"photo"`
	Email        string    `json:"email,omitempty" bson:"email,omitempty"`
	Experience   string    `json:"experience,omitempty" bson:"experience,omitempty"`
	Skills       string    `json:"skills,omitempty" bson:"skills,omitempty"`
	Availability string    `json:"availability,omitempty" bson:"availability,omitempty"`
	ExpectedWage string    `json:"expected_wage,omitempty" bson:"expected_wage,omitempty"`
	Languages    string    `json:"languages,omitempty" bson:"languages,omitempty"`
	Documents    []string  `json:"documents,omitempty" bson:"documents,omitempty"`
	CreatedAt    time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time `json:"updated_at,omitempty" bson:"updated_at,omitempty"`
}
