package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Report struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ReportedBy  string             `json:"reportedBy"  bson:"reportedBy"`
	TargetID    string             `json:"targetId"    bson:"targetId"`
	TargetType  string             `json:"targetType"  bson:"targetType"`
	Reason      string             `json:"reason"      bson:"reason"`
	Notes       string             `json:"notes,omitempty"      bson:"notes,omitempty"`
	Status      string             `json:"status"      bson:"status"`
	ReviewedBy  string             `json:"reviewedBy,omitempty"  bson:"reviewedBy,omitempty"`
	ReviewNotes string             `json:"reviewNotes,omitempty" bson:"reviewNotes,omitempty"`
	CreatedAt   time.Time          `json:"createdAt"   bson:"createdAt"`
	UpdatedAt   time.Time          `json:"updatedAt"   bson:"updatedAt"`

	// New fields for parent reference
	ParentType string `json:"parentType,omitempty" bson:"parentType,omitempty"`
	ParentID   string `json:"parentId,omitempty"   bson:"parentId,omitempty"`

	// New field to indicate whether the reporter has been notified
	Notified bool `json:"notified" bson:"notified"`
}

// // MarshalJSON implements a custom JSON marshaller so that “id” is the hex string of ObjectID.
// func (r *Report) MarshalJSON() ([]byte, error) {
// 	type Alias Report
// 	return json.Marshal(&struct {
// 		ID string `json:"id"`
// 		*Alias
// 	}{
// 		ID:    r.ID.Hex(),
// 		Alias: (*Alias)(r),
// 	})
// }
