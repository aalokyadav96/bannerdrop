package models

type Config struct {
	EntityType string              `bson:"entityType"`
	EntityId   string              `bson:"entityId"`
	Admins     []string            `bson:"admins"` // array of userIds
	SlotGroups map[string][]string `bson:"slotGroups"`
	SlotMeta   map[string]SlotMeta `bson:"slotMeta"`
}

type SlotMeta struct {
	Label     string   `bson:"label,omitempty"`
	Price     int      `bson:"price,omitempty"`
	Location  string   `bson:"location,omitempty"` // e.g., "Table 5", "Lot A12"
	TimeStart int64    `bson:"timeStart,omitempty"`
	TimeEnd   int64    `bson:"timeEnd,omitempty"`
	Tags      []string `bson:"tags,omitempty"` // e.g., ["vip","balcony","haircut"]
}

type Booking struct {
	EntityType string                 `bson:"entityType"`
	EntityId   string                 `bson:"entityId"`
	Group      string                 `bson:"group"`
	SlotId     string                 `bson:"slotId"`
	UserId     string                 `bson:"userId"`
	Metadata   map[string]interface{} `bson:"metadata,omitempty"`
	Timestamp  int64                  `bson:"timestamp"`
}
