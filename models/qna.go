package models

import "time"

type Question struct {
	ID          string    `json:"id" bson:"_id,omitempty"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Tags        []string  `json:"tags"`
	User        string    `json:"user"`
	UserId      string    `json:"userId"`
	Timestamp   time.Time `json:"timestamp"`
}

type Answer struct {
	ID        string    `json:"id" bson:"_id,omitempty"`
	PostID    string    `json:"postid"`
	User      string    `json:"user"`
	Content   string    `json:"content"`
	Votes     int       `json:"votes"`
	Downvotes int       `json:"downvotes"`
	Timestamp time.Time `json:"timestamp"`
	Replies   []string  `json:"replies"`
	IsBest    bool      `json:"isBest"`
}

type Reply struct {
	ID        string    `json:"id" bson:"_id,omitempty"`
	AnswerID  string    `json:"answerId"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}
