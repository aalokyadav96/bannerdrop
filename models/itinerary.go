package models

// Itinerary represents the travel itinerary
type Itinerary struct {
	ItineraryID string  `json:"itineraryid" bson:"itineraryid,omitempty"`
	UserID      string  `json:"user_id" bson:"user_id"`
	Name        string  `json:"name" bson:"name"`
	Description string  `json:"description" bson:"description"`
	StartDate   string  `json:"start_date" bson:"start_date"`
	EndDate     string  `json:"end_date" bson:"end_date"`
	Status      string  `json:"status" bson:"status"` // Draft/Confirmed
	Published   bool    `json:"published" bson:"published"`
	ForkedFrom  *string `json:"forked_from,omitempty" bson:"forked_from,omitempty"`
	Deleted     bool    `json:"-" bson:"deleted,omitempty"` // Internal use only
	// the new day-by-day schedule
	Days []Day `json:"days" bson:"days"`
}

// add these at the top, just below package declaration
type Visit struct {
	Location  string `json:"location" bson:"location"`
	StartTime string `json:"start_time" bson:"start_time"`
	EndTime   string `json:"end_time" bson:"end_time"`
	// nil for the very first visit of a day
	Transport *string `json:"transport,omitempty" bson:"transport,omitempty"`
}

type Day struct {
	Date   string  `json:"date" bson:"date"`
	Visits []Visit `json:"visits" bson:"visits"`
}
