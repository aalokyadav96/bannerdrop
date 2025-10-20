package mq

import (
	"context"
	"encoding/json"
	"log"
	"naevis/db"
	"naevis/rdx"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

/*
----------Hashtags--------------
*/

type HashtagEvent struct {
	EntityType  string `json:"entity_type"`
	EntityID    string `json:"entity_id"`
	HashtagName string `json:"hashtag_name"`
}

// EmitHashtagEvents publishes one event per hashtag (to a Redis Pub/Sub channel).
func EmitHashtagEvent(tagType string, postID string, hashtags []string) {
	ctx := context.Background()
	for _, tag := range hashtags {
		evt := HashtagEvent{
			EntityType:  tagType,
			EntityID:    postID,
			HashtagName: tag,
		}

		data, err := json.Marshal(evt)
		if err != nil {
			log.Printf("[EmitHashtagEvent] marshal error for %s: %v", tag, err)
			continue
		}

		log.Println(evt)

		if err := rdx.Conn.Publish(ctx, "hashtag-events", data).Err(); err != nil {
			log.Printf("[EmitHashtagEvent] publish error for %s: %v", tag, err)
		} else {
			log.Printf("[EmitHashtagEvent] published hashtag event: %+v", evt)
		}
	}
}

func StartHashtagWorker() {
	ctx := context.Background()
	sub := rdx.Conn.Subscribe(ctx, "hashtag-events")
	ch := sub.Channel()

	log.Println("[HashtagWorker] Listening for hashtag events...")

	for msg := range ch {
		var evt HashtagEvent
		if err := json.Unmarshal([]byte(msg.Payload), &evt); err != nil {
			log.Printf("[HashtagWorker] unmarshal error: %v", err)
			continue
		}

		filter := bson.M{
			"name":       evt.HashtagName,
			"entitytype": evt.EntityType,
		}

		// Check if this post is already linked
		existsFilter := bson.M{
			"name":       evt.HashtagName,
			"entitytype": evt.EntityType,
			"posts":      evt.EntityID,
		}
		count, err := db.HashtagCollection.CountDocuments(ctx, existsFilter)
		if err != nil {
			log.Printf("[HashtagWorker] count error: %v", err)
			continue
		}

		update := bson.M{
			"$set": bson.M{"updatedat": time.Now()},
			"$addToSet": bson.M{
				"posts": evt.EntityID,
			},
		}
		if count == 0 {
			update["$inc"] = bson.M{"totalposts": 1}
		}

		opts := options.Update().SetUpsert(true)
		if _, err := db.HashtagCollection.UpdateOne(ctx, filter, update, opts); err != nil {
			log.Printf("[HashtagWorker] update error: %v", err)
			continue
		}

		log.Printf("[HashtagWorker] processed hashtag: %s for entity %s (%s)", evt.HashtagName, evt.EntityID, evt.EntityType)
	}
}
