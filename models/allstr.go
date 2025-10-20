package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	// ID          string    `json:"-" bson:"_id,omitempty"`
	UserID       string    `json:"userid" bson:"userid"`
	Username     string    `json:"username" bson:"username"`
	Email        string    `json:"email" bson:"email"`
	Password     string    `json:"-" bson:"password"`
	PasswordHash string    `json:"password_hash" bson:"password_hash"`
	Role         []string  `json:"role" bson:"role"`
	Name         string    `json:"name,omitempty" bson:"name,omitempty"`
	CreatedAt    time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" bson:"updated_at"`
	Bio          string    `json:"bio,omitempty" bson:"bio,omitempty"`
	Online       bool      `json:"online"`
	LastLogin    time.Time `json:"last_login" bson:"last_login"`
	Avatar       string    `json:"avatar" bson:"avatar"`
	Banner       string    `json:"banner" bson:"banner"`
	ProfileViews int       `json:"profile_views,omitempty" bson:"profile_views,omitempty"`
	PhoneNumber  string    `json:"phone_number,omitempty" bson:"phone_number,omitempty"`
	Address      string    `json:"address,omitempty" bson:"address,omitempty"`
	// DateOfBirth    time.Time         `json:"dob" bson:"dob"`
	SocialLinks    map[string]string `json:"social_links,omitempty" bson:"social_links,omitempty"`
	IsVerified     bool              `json:"is_verified" bson:"is_verified"`
	EmailVerified  bool              `json:"email_verified" bson:"email_verified"`
	FollowersCount int               `json:"followerscount" bson:"followerscount"`
	FollowingCount int               `json:"followscount" bson:"followscount"`
	WalletBalance  float64           `bson:"wallet_balance" json:"wallet_balance"`
	RefreshToken   string            `json:"-" bson:"refreshtoken,omitempty"`
	RefreshExpiry  time.Time         `json:"refreshexp" bson:"refreshexp"`
}

// UserProfileResponse defines the structure for the user profile response
type UserProfileResponse struct {
	UserID         string            `json:"userid" bson:"userid"`
	Username       string            `json:"username" bson:"username"`
	Name           string            `json:"name" bson:"name"`
	Email          string            `json:"email" bson:"email"`
	Bio            string            `json:"bio,omitempty" bson:"bio,omitempty"`
	PhoneNumber    string            `json:"phone_number,omitempty" bson:"phone_number,omitempty"`
	Avatar         string            `json:"avatar" bson:"avatar"`
	Banner         string            `json:"banner" bson:"banner"`
	IsFollowing    bool              `json:"is_following" bson:"is_following"` // Added here
	FollowersCount int               `json:"followerscount" bson:"followerscount"`
	FollowingCount int               `json:"followscount" bson:"followscount"`
	SocialLinks    map[string]string `json:"social_links,omitempty" bson:"social_links,omitempty"`
	Online         bool              `json:"online,omitempty"`
	LastLogin      time.Time         `json:"last_login" bson:"last_login"`
}

type UserFollow struct {
	UserID    string   `json:"userid" bson:"userid"`
	Follows   []string `json:"follows,omitempty" bson:"follows,omitempty"`
	Followers []string `json:"followers,omitempty" bson:"followers,omitempty"`
}

type UserSubscribe struct {
	UserID      string   `json:"userid" bson:"userid"`
	Subscribed  []string `json:"subscribed,omitempty" bson:"subscribed,omitempty"`   // users this user is subscribed to
	Subscribers []string `json:"subscribers,omitempty" bson:"subscribers,omitempty"` // users who subscribed to this user
}

type UserData struct {
	UserID     string `json:"userid" bson:"userid"`
	EntityID   string `json:"entity_id" bson:"entity_id"`
	EntityType string `json:"entity_type" bson:"entity_type"`
	ItemID     string `json:"item_id" bson:"item_id"`
	ItemType   string `json:"item_type" bson:"item_type"`
	CreatedAt  string `json:"created_at" bson:"created_at"`
}

type Response struct {
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

type Setting struct {
	Type        string `json:"type"`
	Value       any    `json:"value"`
	Description string `json:"description"`
}

//	type UserSettings struct {
//		UserID   string    `bson:"userID" json:"userID"`
//		Settings []Setting `bson:"settings" json:"settings"`
//	}
type FeedPost struct {
	Username    string `bson:"username" json:"username"`
	PostID      string `bson:"postid,omitempty" json:"postid"`
	UserID      string `bson:"userid" json:"userid"`
	Type        string `bson:"type" json:"type"`
	Text        string `bson:"text,omitempty" json:"text,omitempty"`
	Title       string `bson:"title,omitempty" json:"title,omitempty"`
	Description string `bson:"description,omitempty" json:"description,omitempty"`
	Caption     string `bson:"caption,omitempty" json:"caption,omitempty"`

	Media       []string          `bson:"media,omitempty" json:"media,omitempty"`             // full file paths (key/filename.extn)
	MediaURL    []string          `bson:"media_url,omitempty" json:"media_url,omitempty"`     // clean filenames
	Thumbnail   string            `bson:"thumbnail,omitempty" json:"thumbnail,omitempty"`     // video thumbnail
	Resolutions []int             `bson:"resolutions,omitempty" json:"resolutions,omitempty"` // optional resolutions
	Subtitles   map[string]string `bson:"subtitles,omitempty" json:"subtitles,omitempty"`     // lang → file path
	Tags        []string          `bson:"tags,omitempty" json:"tags,omitempty"`               // hashtags or topics

	Timestamp string               `bson:"timestamp" json:"timestamp"`
	CreatedAt time.Time            `bson:"created_at" json:"created_at"`
	Likes     int64                `bson:"likes" json:"likes"`
	Likers    []primitive.ObjectID `bson:"likers,omitempty" json:"likers,omitempty"`

	// Legacy / optional field kept for backward compatibility
	Content string `bson:"content,omitempty" json:"content,omitempty"`
}

type Activity struct {
	// Username     string              `json:"username,omitempty" bson:"username,omitempty"`
	PlaceID      string    `json:"placeId,omitempty" bson:"placeId,omitempty"`
	Action       string    `json:"action,omitempty" bson:"action,omitempty"`
	PerformedBy  string    `json:"performedBy,omitempty" bson:"performedBy,omitempty"`
	Timestamp    time.Time `json:"timestamp,omitempty" bson:"timestamp,omitempty"`
	Details      string    `json:"details,omitempty" bson:"details,omitempty"`
	IPAddress    string    `json:"ipAddress,omitempty" bson:"ipAddress,omitempty"`
	DeviceInfo   string    `json:"deviceInfo,omitempty" bson:"deviceInfo,omitempty"`
	ActivityID   string    `json:"activityid" bson:"activityid,omitempty"`
	UserID       string    `json:"user_id" bson:"user_id"`
	ActivityType string    `json:"activity_type" bson:"activity_type"` // e.g., "follow", "review", "buy"
	EntityID     string    `json:"entity_id,omitempty" bson:"entity_id,omitempty"`
	EntityType   *string   `json:"entity_type,omitempty" bson:"entity_type,omitempty"` // "event", "place", or null
}
type Merch struct {
	MerchID string `json:"merchid" bson:"merchid"`
	// EventID     string             `json:"eventid" bson:"eventid"` // Reference to Event ID
	Name        string             `json:"name" bson:"name"`
	Slug        string             `json:"slug,omitempty" bson:"slug,omitempty"`         // URL-friendly name (e.g. "concert-tshirt")
	SKU         string             `json:"sku,omitempty" bson:"sku,omitempty"`           // Stock Keeping Unit, unique per product
	Category    string             `json:"category,omitempty" bson:"category,omitempty"` // e.g. “T-Shirts”, “Accessories”
	Price       float64            `json:"price" bson:"price"`
	Discount    float64            `json:"discount,omitempty" bson:"discount,omitempty"`         // e.g. 0.10 for 10% off
	Stock       int                `json:"stock" bson:"stock"`                                   // Number of items available
	StockStatus string             `json:"stock_status,omitempty" bson:"stock_status,omitempty"` // e.g. “In Stock”, “Out of Stock”, “Preorder”
	MerchPhoto  string             `json:"merch_pic" bson:"merch_pic"`
	Gallery     []string           `json:"gallery,omitempty" bson:"gallery,omitempty"` // Additional image filenames
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	EntityID    string             `json:"entity_id" bson:"entity_id"`
	EntityType  string             `json:"entity_type" bson:"entity_type"` // “event” or “place”
	Description string             `json:"description,omitempty" bson:"description,omitempty"`
	ShortDesc   string             `json:"short_desc,omitempty" bson:"short_desc,omitempty"` // One-line summary
	Rating      float64            `json:"rating,omitempty" bson:"rating,omitempty"`         // Average rating (0.0–5.0)
	ReviewCount int                `json:"review_count,omitempty" bson:"review_count,omitempty"`
	Weight      float64            `json:"weight,omitempty" bson:"weight,omitempty"`         // In kilograms/pounds
	Dimensions  string             `json:"dimensions,omitempty" bson:"dimensions,omitempty"` // e.g. “30×20×2 cm”
	Tags        []string           `json:"tags,omitempty" bson:"tags,omitempty"`             // e.g. ["rock", "tshirt"]
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updatedAt"`
	UserID      primitive.ObjectID `bson:"user_id" json:"userId"`
}

type Menu struct {
	MenuID      string             `json:"menuid" bson:"menuid"`
	PlaceID     string             `json:"placeid" bson:"placeid"` // Reference to Place ID
	Name        string             `json:"name" bson:"name"`
	Price       float64            `json:"price" bson:"price"`
	Stock       int                `json:"stock" bson:"stock"` // Number of items available
	MenuPhoto   string             `json:"menu_pic" bson:"menu_pic"`
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Description string             `json:"description,omitempty" bson:"description,omitempty"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UserID      primitive.ObjectID `bson:"user_id" json:"userId"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updatedAt"`
}

type Ticket struct {
	TicketID    string             `json:"ticketid" bson:"ticketid"`
	EventID     string             `json:"eventid" bson:"eventid"`
	Name        string             `json:"name" bson:"name"`
	Price       float64            `json:"price" bson:"price"`
	Currency    string             `json:"currency" bson:"currency"`
	Color       string             `json:"color" bson:"color"`
	Quantity    int                `json:"quantity" bson:"quantity"`
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	EntityID    string             `json:"entity_id" bson:"entity_id"`
	EntityType  string             `json:"entity_type" bson:"entity_type"` // "event" or "place"
	Available   int                `json:"available" bson:"available"`
	Total       int                `json:"total" bson:"total"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	Description string             `bson:"description,omitempty" json:"description"`
	Sold        int                `bson:"sold" json:"sold"`
	SeatStart   int                `bson:"seatstart" json:"seatstart"`
	SeatEnd     int                `bson:"seatend" json:"seatend"`
	Seats       []string           `bson:"seats" json:"seats"` // 👈 new field
	UpdatedAt   time.Time          `bson:"updated_at" json:"updatedAt"`
}

type Seat struct {
	ID         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	EntityID   primitive.ObjectID `json:"entity_id" bson:"entity_id"`
	EntityType string             `json:"entity_type" bson:"entity_type"` // e.g., "event" or "place"
	SeatNumber string             `json:"seat_number" bson:"seat_number"`
	UserID     primitive.ObjectID `json:"user_id" bson:"user_id,omitempty"`
	Status     string             `json:"status" bson:"status"` // e.g., "booked", "available"
}

// UserProfileResponse defines the structure for the user profile response
type UserSuggest struct {
	Username    string `json:"username" bson:"username"`
	UserID      string `json:"userid" bson:"userid"`
	IsFollowing bool
	Bio         string `json:"bio,omitempty" bson:"bio,omitempty"`
}

type Suggestion struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Type        string             `json:"type" bson:"type"` // e.g., "place" or "event"
	Title       string             `json:"title" bson:"title"`
	Description string             `json:"description,omitempty" bson:"description,omitempty"`
	Name        string             `json:"name"`
}

type Review struct {
	EntityID    string    `json:"entity_id" bson:"entity_id"`
	EntityType  string    `json:"entity_type" bson:"entity_type"` // "event" or "place"
	Comment     string    `json:"comment,omitempty" bson:"comment,omitempty"`
	UpdatedAt   time.Time `json:"updated_at" bson:"updated_at"`
	Content     string    `bson:"content" json:"content"`
	ReviewID    string    `json:"reviewid" bson:"reviewid"`
	UserID      string    `json:"userid" bson:"userid"` // Reference to User ID
	Rating      int       `json:"rating" bson:"rating"` // Rating out of 5
	Date        time.Time `json:"date" bson:"date"`     // Date of the review
	Likes       int       `json:"likes,omitempty" bson:"likes,omitempty"`
	Dislikes    int       `json:"dislikes,omitempty" bson:"dislikes,omitempty"`
	Attachments []string  `json:"attachments,omitempty" bson:"attachments,omitempty"`
	CreatedAt   string    `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
}

type Media struct {
	MediaID       string             `json:"mediaid" bson:"mediaid"`
	Type          string             `json:"type" bson:"type"` // "image", "video", "text"
	URL           string             `json:"url,omitempty" bson:"url,omitempty"`
	ThumbnailURL  string             `json:"thumbnailUrl,omitempty" bson:"thumbnailUrl,omitempty"`
	Caption       string             `json:"caption,omitempty" bson:"caption,omitempty"`
	Description   string             `json:"description,omitempty" bson:"description,omitempty"`
	CreatorID     string             `json:"creatorid" bson:"creatorid"`
	LikesCount    int                `json:"likesCount" bson:"likesCount"`
	CommentsCount int                `json:"commentsCount" bson:"commentsCount"`
	Visibility    string             `json:"visibility,omitempty" bson:"visibility,omitempty"`
	Tags          []string           `json:"tags,omitempty" bson:"tags,omitempty"` // e.g., song:123, event:456
	Duration      float64            `json:"duration,omitempty" bson:"duration,omitempty"`
	FileSize      int64              `json:"fileSize,omitempty" bson:"fileSize,omitempty"`
	MimeType      string             `json:"mimeType,omitempty" bson:"mimeType,omitempty"`
	IsFeatured    bool               `json:"isFeatured,omitempty" bson:"isFeatured,omitempty"`
	EntityID      string             `json:"entityid" bson:"entityid"`
	EntityType    string             `json:"entitytype" bson:"entitytype"` // "event", "place", etc.
	CreatedAt     time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt     time.Time          `json:"updatedAt" bson:"updatedAt"`
	UserID        primitive.ObjectID `json:"userId" bson:"user_id"`
}

type Place struct {
	PlaceID           string            `json:"placeid" bson:"placeid"`
	Name              string            `json:"name" bson:"name"`
	ShortDesc         string            `json:"short_desc" bson:"short_desc"`
	Description       string            `json:"description" bson:"description"`
	Place             string            `json:"place" bson:"place"`
	Capacity          int               `json:"capacity" bson:"capacity"`
	Date              time.Time         `json:"date" bson:"date"`
	Address           string            `json:"address" bson:"address"`
	CreatedBy         string            `json:"createdBy,omitempty" bson:"createdBy,omitempty"`
	OrganizerName     string            `json:"organizer_name" bson:"organizer_name"`
	OrganizerContact  string            `json:"organizer_contact" bson:"organizer_contact"`
	Category          string            `json:"category" bson:"category"`
	Banner            string            `json:"banner" bson:"banner"`
	WebsiteURL        string            `json:"website_url" bson:"website_url"`
	Status            string            `json:"status" bson:"status"`
	AccessibilityInfo string            `json:"accessibility_info" bson:"accessibility_info"`
	SocialMediaLinks  []string          `json:"social_links" bson:"social_links"`
	Tags              []string          `json:"tags" bson:"tags"`
	CustomFields      map[string]any    `json:"custom_fields" bson:"custom_fields"`
	CreatedAt         time.Time         `json:"created_at" bson:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at" bson:"updated_at"`
	City              string            `json:"city,omitempty" bson:"city,omitempty"`
	Country           string            `json:"country,omitempty" bson:"country,omitempty"`
	ZipCode           string            `json:"zipCode,omitempty" bson:"zipCode,omitempty"`
	Jobs              string            `json:"jobs,omitempty" bson:"jobs,omitempty"`
	Location          Coordinates       `json:"location" bson:"location,omitempty"`
	Phone             string            `json:"phone,omitempty" bson:"phone,omitempty"`
	Website           string            `json:"website,omitempty" bson:"website,omitempty"`
	IsOpen            bool              `json:"isopen,omitempty" bson:"isopen,omitempty"`
	Distance          float64           `json:"distance,omitempty" bson:"distance,omitempty"`
	Views             int               `json:"views,omitempty" bson:"views,omitempty"`
	ReviewCount       int               `json:"reviewcount,omitempty" bson:"reviewcount,omitempty"`
	SocialLinks       map[string]string `json:"socialLinks,omitempty" bson:"socialLinks,omitempty"`
	UpdatedBy         string            `json:"updatedBy,omitempty" bson:"updatedBy,omitempty"`
	DeletedAt         *time.Time        `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	Amenities         []string          `json:"amenities,omitempty" bson:"amenities,omitempty"`
	Events            []string          `json:"events,omitempty" bson:"events,omitempty"`
	OperatingHours    []string          `json:"operatinghours,omitempty" bson:"operatinghours,omitempty"`
	Keywords          []string          `json:"keywords,omitempty" bson:"keywords,omitempty"`
}

type PlaceStatus string

const (
	Active   PlaceStatus = "active"
	Inactive PlaceStatus = "inactive"
	Closed   PlaceStatus = "closed"
)

type Coordinates struct {
	Latitude  float64 `json:"latitude,omitempty" bson:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty" bson:"longitude,omitempty"`
}

type CheckIn struct {
	UserID    string    `json:"userId,omitempty" bson:"userId,omitempty"`
	PlaceID   string    `json:"placeId,omitempty" bson:"placeId,omitempty"`
	Timestamp time.Time `json:"timestamp,omitempty" bson:"timestamp,omitempty"`
	Comment   string    `json:"comment,omitempty" bson:"comment,omitempty"`
	Rating    float64   `json:"rating,omitempty" bson:"rating,omitempty"` // Optional
	Medias    []Media   `json:"images,omitempty" bson:"images,omitempty"` // Optional
}

type PlaceVersion struct {
	PlaceID   string            `json:"placeId,omitempty" bson:"placeId,omitempty"`
	Version   int               `json:"version,omitempty" bson:"version,omitempty"`
	Data      Place             `json:"data,omitempty" bson:"data,omitempty"`
	UpdatedAt time.Time         `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
	UpdatedBy string            `json:"updatedBy,omitempty" bson:"updatedBy,omitempty"`
	Changes   map[string]string `json:"changes,omitempty" bson:"changes,omitempty"`
}

type OperatingHours struct {
	Day          []string `json:"day,omitempty" bson:"day,omitempty"`
	OpeningHours []string `json:"opening,omitempty" bson:"opening,omitempty"`
	ClosingHours []string `json:"closing,omitempty" bson:"closing,omitempty"`
	TimeZone     string   `json:"timeZone,omitempty" bson:"timeZone,omitempty"`
}

type Tag struct {
	ID     string   `json:"id,omitempty" bson:"_id,omitempty"`
	Name   string   `json:"name,omitempty" bson:"name,omitempty"`
	Places []string `json:"places,omitempty" bson:"places,omitempty"` // List of Place IDs tagged with this keyword
}

const (
	PlaceStatusActive     = "active"
	PlaceStatusClosed     = "closed"
	PlaceStatusRenovation = "under renovation"
)

const (
	MediaTypeImage    = "image"
	MediaTypeVideo    = "video"
	MediaTypePhoto360 = "photo360"
)

type Business struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name        string             `json:"name" bson:"name"`
	Type        string             `json:"type" bson:"type"`
	Location    string             `json:"location" bson:"location"`
	Description string             `json:"description" bson:"description"`
}

type MenuItem struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name        string             `json:"name" bson:"name"`
	Price       float64            `json:"price" bson:"price"`
	Description string             `json:"description" bson:"description"`
	CreatedAt   time.Time          `json:"createdAt"`
}

type Promotion struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Title       string             `json:"title" bson:"title"`
	Description string             `json:"description" bson:"description"`
	ExpiryDate  time.Time          `json:"expiry_date" bson:"expiry_date"`
}

// Owner Management Handlers
type Owner struct {
	ID       primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name     string             `json:"name" bson:"name"`
	Email    string             `json:"email" bson:"email"`
	Password string             `json:"password" bson:"password"`
}

type Event struct {
	EventID          string      `json:"eventid" bson:"eventid"`
	Title            string      `json:"title" bson:"title"`
	Description      string      `json:"description" bson:"description"`
	Date             time.Time   `json:"date" bson:"date"`
	PlaceID          string      `json:"placeid" bson:"placeid"`
	PlaceName        string      `json:"placename" bson:"placename"`
	Location         string      `json:"location" bson:"location"`
	Coords           Coordinates `json:"coords" bson:"coords"`
	CreatorID        string      `json:"creatorid" bson:"creatorid"`
	Tickets          []Ticket    `json:"tickets" bson:"tickets"`
	Merch            []Merch     `json:"merch" bson:"merch"`
	StartDateTime    time.Time   `json:"start_date_time" bson:"start_date_time"`
	EndDateTime      time.Time   `json:"end_date_time" bson:"end_date_time"`
	Category         string      `json:"category" bson:"category"`
	Banner           string      `json:"banner" bson:"banner"`
	SeatingPlanImage string      `json:"seating" bson:"seating"`
	WebsiteURL       string      `json:"website_url" bson:"website_url"`
	Status           string      `json:"status" bson:"status"`
	Tags             []string    `json:"tags" bson:"tags"`
	CreatedAt        time.Time   `json:"created_at" bson:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at" bson:"updated_at"`
	FAQs             []FAQ       `json:"faqs" bson:"faqs"`
	OrganizerName    string      `json:"organizer_name" bson:"organizer_name"`
	OrganizerContact string      `json:"organizer_contact" bson:"organizer_contact"`
	Artists          []string    `json:"artists,omitempty" bson:"artists,omitempty"`
	Published        string      `json:"published,omitempty" bson:"published,omitempty"`

	// Computed fields for frontend filters
	Prices   []float64 `json:"prices,omitempty" bson:"-"`
	Currency string    `json:"currency,omitempty" bson:"-"`
}

// FAQ represents a single FAQ structure
type FAQ struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type SocialMediaLinks struct {
	Title string `json:"title"`
	Url   string `json:"Url"`
}

type PurchasedTicket struct {
	EventID      string
	TicketID     string
	UserID       string
	BuyerName    string
	UniqueCode   string
	PurchaseDate time.Time
}
