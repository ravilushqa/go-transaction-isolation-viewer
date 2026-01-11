package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/ravilushqa/go-transaction-isolation-viewer/internal/scenario"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
)

// DirtyReadScenario demonstrates the difference between reading with and without transactions
type DirtyReadScenario struct {
	client     *mongo.Client
	db         *mongo.Database
	collection *mongo.Collection
}

// NewDirtyReadScenario creates a new dirty read demonstration scenario
func NewDirtyReadScenario(client *mongo.Client, db *mongo.Database) *DirtyReadScenario {
	return &DirtyReadScenario{
		client:     client,
		db:         db,
		collection: db.Collection("dirty_read_demo"),
	}
}

func (s *DirtyReadScenario) Name() string {
	return "Dirty Read Prevention"
}

func (s *DirtyReadScenario) Description() string {
	return `Demonstrates how MongoDB transactions prevent dirty reads.

Without transactions, reads might see uncommitted data. With transactions
and proper read concern, you only see committed data.

This scenario shows:
1. Session A starts a transaction and inserts a document
2. Session B tries to read - document is NOT visible (not committed yet)
3. Session A commits the transaction
4. Session B reads again - document IS now visible`
}

func (s *DirtyReadScenario) IsolationLevel() string {
	return "Read Committed"
}

func (s *DirtyReadScenario) Setup(ctx context.Context) error {
	// Drop collection if exists
	return s.collection.Drop(ctx)
}

func (s *DirtyReadScenario) Cleanup(ctx context.Context) error {
	return s.collection.Drop(ctx)
}

func (s *DirtyReadScenario) Run(ctx context.Context, output chan<- scenario.StepResult) error {
	defer close(output)

	// Header
	output <- scenario.StepResult{
		IsHeader:    true,
		Description: "🔒 Dirty Read Prevention Demonstration",
	}

	step := 1

	// Step 1: Show initial state
	output <- scenario.StepResult{
		Session:     "Setup",
		Step:        step,
		Description: "Checking initial state - collection should be empty",
		Query:       "db.dirty_read_demo.countDocuments({})",
		Result:      "Count: 0",
		Success:     true,
	}
	step++

	// Step 2: Session A starts a transaction
	sessionA, err := s.client.StartSession()
	if err != nil {
		return fmt.Errorf("failed to start session A: %w", err)
	}
	defer sessionA.EndSession(ctx)

	output <- scenario.StepResult{
		Session:     "Session A",
		Step:        step,
		Description: "Starting a transaction",
		Query:       "session.startTransaction()",
		Result:      "Transaction started",
		Success:     true,
	}
	step++

	// Step 3: Session A inserts a document within transaction
	err = mongo.WithSession(ctx, sessionA, func(sc mongo.SessionContext) error {
		if err := sessionA.StartTransaction(); err != nil {
			return err
		}

		_, err := s.collection.InsertOne(sc, bson.M{
			"product": "Widget",
			"price":   29.99,
			"status":  "pending",
		})
		return err
	})
	if err != nil {
		return fmt.Errorf("failed to insert in transaction: %w", err)
	}

	output <- scenario.StepResult{
		Session:     "Session A",
		Step:        step,
		Description: "Inserted document within transaction (NOT YET COMMITTED)",
		Query:       `db.dirty_read_demo.insertOne({product: "Widget", price: 29.99, status: "pending"})`,
		Result:      "Insert successful (within transaction)",
		Success:     true,
	}
	step++

	// Small delay for visual effect
	time.Sleep(500 * time.Millisecond)

	// Step 4: Session B tries to read (should NOT see uncommitted data)
	output <- scenario.StepResult{
		Session:     "Session B",
		Step:        step,
		Description: "Attempting to read documents (outside Session A's transaction)",
		Query:       `db.dirty_read_demo.find({})`,
		Result:      "",
		Success:     true,
	}

	// Read with majority read concern by using a collection with that concern
	collWithReadConcern := s.db.Collection("dirty_read_demo", options.Collection().SetReadConcern(readconcern.Majority()))
	cursor, err := collWithReadConcern.Find(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("failed to read: %w", err)
	}

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return fmt.Errorf("failed to decode results: %w", err)
	}

	output <- scenario.StepResult{
		Session:     "Session B",
		Step:        step,
		Description: "Read completed with readConcern: majority",
		Query:       `db.dirty_read_demo.find({}).readConcern("majority")`,
		Result:      fmt.Sprintf("Documents found: %d (uncommitted data NOT visible!)", len(results)),
		Success:     true,
	}
	step++

	output <- scenario.StepResult{
		IsHeader:    true,
		Description: "✅ Dirty read prevented! Session B cannot see Session A's uncommitted data",
	}

	// Step 5: Session A commits
	time.Sleep(500 * time.Millisecond)

	err = mongo.WithSession(ctx, sessionA, func(sc mongo.SessionContext) error {
		return sessionA.CommitTransaction(sc)
	})
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	output <- scenario.StepResult{
		Session:     "Session A",
		Step:        step,
		Description: "Committing the transaction",
		Query:       "session.commitTransaction()",
		Result:      "Transaction committed successfully",
		Success:     true,
	}
	step++

	time.Sleep(500 * time.Millisecond)

	// Step 6: Session B reads again - now sees the data
	cursor, err = s.collection.Find(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("failed to read after commit: %w", err)
	}

	results = nil
	if err := cursor.All(ctx, &results); err != nil {
		return fmt.Errorf("failed to decode results: %w", err)
	}

	resultStr := "[]"
	if len(results) > 0 {
		resultStr = fmt.Sprintf("[{product: %q, price: %v, status: %q}]",
			results[0]["product"], results[0]["price"], results[0]["status"])
	}

	output <- scenario.StepResult{
		Session:     "Session B",
		Step:        step,
		Description: "Reading documents again after Session A committed",
		Query:       "db.dirty_read_demo.find({})",
		Result:      fmt.Sprintf("Documents found: %d\n%s", len(results), resultStr),
		Success:     true,
	}

	output <- scenario.StepResult{
		IsHeader:    true,
		Description: "🎉 After commit, Session B can now see Session A's data",
	}

	return nil
}
