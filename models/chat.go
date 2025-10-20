package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Message struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	ChatID    string             `bson:"chatID" json:"chatID"`
	UserID    string             `bson:"userID" json:"userID"`
	Text      string             `bson:"text,omitempty" json:"text,omitempty"`
	FileURL   string             `bson:"fileURL,omitempty" json:"fileURL,omitempty"`
	FileType  string             `bson:"fileType,omitempty" json:"fileType,omitempty"` // "image" or "video"
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	ReplyTo   *ReplyRef          `bson:"replyTo,omitempty" json:"replyTo,omitempty"`
}

type Chat struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	Users       []string           `bson:"users" json:"users"`
	LastMessage MessagePreview     `bson:"lastMessage" json:"lastMessage"`
	ReadStatus  map[string]bool    `bson:"readStatus,omitempty" json:"readStatus,omitempty"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type MessagePreview struct {
	Text      string    `bson:"text" json:"text"`
	SenderID  string    `bson:"senderId" json:"senderId"`
	Timestamp time.Time `bson:"timestamp" json:"timestamp"`
}

// ReplyRef represents the clientâ€side â€œreplyToâ€ payload.
type ReplyRef struct {
	ID   string `json:"id"`
	User string `json:"user"`
	Text string `json:"text"`
}

type Like struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	UserID     string             `bson:"user_id"`
	EntityType string             `bson:"entity_type"` // e.g. "post"
	EntityID   string             `bson:"entity_id"`   // e.g. post ID
	CreatedAt  time.Time          `bson:"created_at"`
}
