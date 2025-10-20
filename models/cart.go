package models

import "time"

// CartItem represents a single item in the user's cart.
type CartItem struct {
	UserID     string    `json:"userId" bson:"userId"`
	Category   string    `json:"category" bson:"category"` // e.g. "crops", "merchandise", "tickets", "tools"
	ItemId     string    `json:"itemId" bson:"itemId"`     // stable identifier for the product/service
	ItemName   string    `json:"itemName" bson:"itemName"` // human-readable name
	ItemType   string    `json:"itemType" bson:"itemType"` // breed,variant et
	Unit       string    `json:"unit,omitempty" bson:"unit,omitempty"`
	EntityId   string    `json:"entityId,omitempty" bson:"entityId,omitempty"`     // e.g. farmId, eventId, shopId
	EntityName string    `json:"entityName,omitempty" bson:"entityName,omitempty"` // e.g. farm name, event title
	EntityType string    `json:"entityType,omitempty" bson:"entityType,omitempty"` // e.g. "farm", "event", "artist"
	Quantity   int       `json:"quantity" bson:"quantity"`
	Price      float64   `json:"price" bson:"price"`     // unit price
	AddedAt    time.Time `json:"addedAt" bson:"addedAt"` // timestamp of when the item was added
}

// CheckoutSession represents a pre-order session, grouped by category.
type CheckoutSession struct {
	UserID    string                `json:"userId" bson:"userId"`
	Items     map[string][]CartItem `json:"items" bson:"items"` // grouped by category
	Address   string                `json:"address" bson:"address"`
	Total     float64               `json:"total" bson:"total"`
	CreatedAt time.Time             `json:"createdAt" bson:"createdAt"`
}

// Order represents a finalized order.
type Order struct {
	OrderID       string                `json:"orderId" bson:"orderId"`
	UserID        string                `json:"userId" bson:"userId"`
	Items         map[string][]CartItem `json:"items" bson:"items"` // grouped by category
	Address       string                `json:"address" bson:"address"`
	PaymentMethod string                `json:"paymentMethod" bson:"paymentMethod"`
	Total         float64               `json:"total" bson:"total"`
	Status        string                `json:"status" bson:"status"` // e.g. "pending", "completed"
	ApprovedBy    []string              `json:"approvedBy" bson:"approvedBy"`
	CreatedAt     time.Time             `json:"createdAt" bson:"createdAt"`
}
