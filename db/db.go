package db

import (
	"context"
	"log"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	Client *mongo.Client
	// Your collections:
	AnalyticsCollection         *mongo.Collection
	AccountsCollection          *mongo.Collection
	AppealsCollection           *mongo.Collection
	MapsCollection              *mongo.Collection
	CartCollection              *mongo.Collection
	OrderCollection             *mongo.Collection
	CatalogueCollection         *mongo.Collection
	FarmsCollection             *mongo.Collection
	FarmOrdersCollection        *mongo.Collection
	CropsCollection             *mongo.Collection
	CommentsCollection          *mongo.Collection
	HashtagCollection           *mongo.Collection
	UserCollection              *mongo.Collection
	TransactionCollection       *mongo.Collection
	LikesCollection             *mongo.Collection
	ProductCollection           *mongo.Collection
	IdempotencyCollection       *mongo.Collection
	ItineraryCollection         *mongo.Collection
	JournalCollection           *mongo.Collection
	UserDataCollection          *mongo.Collection
	TicketsCollection           *mongo.Collection
	BehindTheScenesCollection   *mongo.Collection
	PurchasedTicketsCollection  *mongo.Collection
	ReviewsCollection           *mongo.Collection
	SettingsCollection          *mongo.Collection
	FollowingsCollection        *mongo.Collection
	PlacesCollection            *mongo.Collection
	SlotCollection              *mongo.Collection
	DateCapsCollection          *mongo.Collection
	BookingsCollection          *mongo.Collection
	PostsCollection             *mongo.Collection
	BlogPostsCollection         *mongo.Collection
	FilesCollection             *mongo.Collection
	MerchCollection             *mongo.Collection
	MenuCollection              *mongo.Collection
	ActivitiesCollection        *mongo.Collection
	EventsCollection            *mongo.Collection
	ArtistEventsCollection      *mongo.Collection
	SongsCollection             *mongo.Collection
	CouponCollection            *mongo.Collection
	MediaCollection             *mongo.Collection
	ArtistsCollection           *mongo.Collection
	ChatsCollection             *mongo.Collection
	MessagesCollection          *mongo.Collection
	ReportsCollection           *mongo.Collection
	RecipeCollection            *mongo.Collection
	BaitoCollection             *mongo.Collection
	ModeratorApplications       *mongo.Collection
	BaitoApplicationsCollection *mongo.Collection
	BaitoWorkerCollection       *mongo.Collection
	TiersCollection             *mongo.Collection
	SearchCollection            *mongo.Collection
	ServiceCollection           *mongo.Collection
	SubscribersCollection       *mongo.Collection
)

// limiter chan to cap concurrent Mongo ops
var mongoLimiter = make(chan struct{}, 100) // allow up to 100 concurrent ops

func init() {
	_ = godotenv.Load()

	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatal("❌ MONGODB_URI environment variable not set")
	}

	clientOpts := options.Client().
		ApplyURI(uri).
		SetMaxPoolSize(100).
		SetMinPoolSize(10).
		SetRetryWrites(true)

	var err error
	Client, err = mongo.Connect(context.Background(), clientOpts)
	if err != nil {
		log.Fatalf("❌ Failed to connect to MongoDB: %v", err)
	}
	if err := Client.Ping(context.Background(), nil); err != nil {
		log.Fatalf("❌ Mongo ping failed: %v", err)
	}

	log.Printf("✅ MongoDB connected (%s) maxPool=%d minPool=%d; Goroutines at start: %d",
		uri, *clientOpts.MaxPoolSize, *clientOpts.MinPoolSize, runtime.NumGoroutine(),
	)

	// Graceful shutdown hook
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		log.Println("🛑 Disconnecting from MongoDB...")
		_ = Client.Disconnect(context.Background())
		os.Exit(0)
	}()

	// Optional: log connection stats periodically
	go logPoolStats()

	// Initialize your collections
	db := Client.Database("eventdb")
	dbx := Client.Database("naevis")
	AccountsCollection = db.Collection("accounts")
	ActivitiesCollection = db.Collection("activities")
	AnalyticsCollection = db.Collection("analytics")
	AppealsCollection = db.Collection("appeals")
	ArtistEventsCollection = db.Collection("artistevents")
	ArtistsCollection = db.Collection("artists")
	BaitoApplicationsCollection = db.Collection("baitoapply")
	BaitoCollection = db.Collection("baitos")
	BaitoWorkerCollection = db.Collection("baitoworkers")
	BlogPostsCollection = db.Collection("blogposts")
	BookingsCollection = db.Collection("bookings")
	BehindTheScenesCollection = db.Collection("bts")
	CartCollection = db.Collection("cart")
	CatalogueCollection = db.Collection("catalogue")
	ChatsCollection = db.Collection("chats")
	CommentsCollection = db.Collection("comments")
	CouponCollection = db.Collection("coupons")
	CropsCollection = db.Collection("crops")
	DateCapsCollection = db.Collection("date_caps")
	EventsCollection = db.Collection("events")
	FarmsCollection = db.Collection("farms")
	PostsCollection = db.Collection("feedposts")
	FilesCollection = db.Collection("files")
	FollowingsCollection = db.Collection("followings")
	FarmOrdersCollection = db.Collection("forders")
	HashtagCollection = db.Collection("hashtags")
	IdempotencyCollection = db.Collection("idempotency")
	ItineraryCollection = db.Collection("itinerary")
	JournalCollection = db.Collection("journals")
	LikesCollection = db.Collection("likes")
	MapsCollection = db.Collection("maps")
	MediaCollection = db.Collection("media")
	MenuCollection = db.Collection("menu")
	MerchCollection = db.Collection("merch")
	MessagesCollection = db.Collection("messages")
	ModeratorApplications = db.Collection("modapps")
	OrderCollection = db.Collection("orders")
	PlacesCollection = db.Collection("places")
	ProductCollection = db.Collection("products")
	PurchasedTicketsCollection = db.Collection("purticks")
	RecipeCollection = db.Collection("recipes")
	ReportsCollection = db.Collection("reports")
	ReviewsCollection = db.Collection("reviews")
	ServiceCollection = db.Collection("service")
	SettingsCollection = db.Collection("settings")
	SlotCollection = db.Collection("slots")
	SongsCollection = db.Collection("songs")
	SubscribersCollection = db.Collection("subscribers")
	TicketsCollection = db.Collection("ticks")
	TiersCollection = db.Collection("tiers")
	TransactionCollection = db.Collection("transactions")
	UserDataCollection = db.Collection("userdata")
	UserCollection = db.Collection("users")
	SearchCollection = dbx.Collection("users")
}

// logPoolStats logs basic goroutine and pool stats every 60s (optional)
func logPoolStats() {
	for {
		time.Sleep(60 * time.Second)
		log.Printf("📊 Mongo Stats: Goroutines=%d | MongoOpsRunning=%d", runtime.NumGoroutine(), len(mongoLimiter))
	}
}

// PingMongo can be used in your /health endpoint
func PingMongo() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return Client.Ping(ctx, nil)
}

// WithMongo wraps any Mongo operation with concurrency and timeout + minimal retry
func WithMongo(op func(ctx context.Context) error) error {
	mongoLimiter <- struct{}{}        // acquire slot
	defer func() { <-mongoLimiter }() // release slot

	var err error
	for i := 0; i < 2; i++ { // 1 retry max
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = op(ctx)
		if err == nil {
			return nil
		}
		log.Printf("⚠️ Mongo op failed: %v (retry %d)", err, i+1)
		time.Sleep(200 * time.Millisecond)
	}
	return err
}

// OptionsFindLatest provides a find option with latest sort
func OptionsFindLatest(limit int64) *options.FindOptions {
	opts := options.Find()
	opts.SetSort(map[string]interface{}{"createdAt": -1})
	opts.SetLimit(limit)
	return opts
}
