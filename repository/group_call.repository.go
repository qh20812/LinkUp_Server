package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type GroupCallDocument struct {
	CallID       string     `bson:"_id"`
	ChatID       string     `bson:"chat_id"`
	CallerID     string     `bson:"caller_id"`
	CallType     string     `bson:"call_type"`
	Participants []string   `bson:"participants"`
	Status       string     `bson:"status"`
	CreatedAt    time.Time  `bson:"created_at"`
	EndedAt      *time.Time `bson:"ended_at,omitempty"`
	UpdatedAt    time.Time  `bson:"updated_at"`
}

type GroupCallRepository struct {
	collection *mongo.Collection
}

func NewGroupCallRepository(db *mongo.Database) *GroupCallRepository {
	return &GroupCallRepository{
		collection: db.Collection("group_calls"),
	}
}

func (r *GroupCallRepository) Create(ctx context.Context, doc *GroupCallDocument) error {
	_, err := r.collection.InsertOne(ctx, doc)
	return err
}

func (r *GroupCallRepository) UpdateParticipants(ctx context.Context, callID string, participants []string) error {
	_, err := r.collection.UpdateOne(ctx,
		bson.M{"_id": callID},
		bson.M{"$set": bson.M{
			"participants": participants,
			"updated_at":   time.Now().UTC(),
		}},
	)
	return err
}

func (r *GroupCallRepository) UpdateStatus(ctx context.Context, callID string, status string, endedAt *time.Time) error {
	setFields := bson.M{"status": status, "updated_at": time.Now().UTC()}
	if endedAt != nil {
		setFields["ended_at"] = endedAt
	}
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": callID}, bson.M{"$set": setFields})
	return err
}
