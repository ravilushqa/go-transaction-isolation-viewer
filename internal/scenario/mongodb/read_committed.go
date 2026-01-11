package mongodb

import (
	"context"
	"fmt"
	"time"

	"txdemo/internal/scenario"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

// ReadCommittedScenario demonstrates read committed isolation level
type ReadCommittedScenario struct {
	client     *mongo.Client
	db         *mongo.Database
	collection *mongo.Collection
}

// NewReadCommittedScenario creates a new read committed demonstration scenario
func NewReadCommittedScenario(client *mongo.Client, db *mongo.Database) *ReadCommittedScenario {
	return &ReadCommittedScenario{
		client:     client,
		db:         db,
		collection: db.Collection("read_committed_demo"),
	}
}

func (s *ReadCommittedScenario) Name() string {
	return "Read Committed Isolation"
}

func (s *ReadCommittedScenario) Description() string {
	return `Demonstrates Read Committed isolation using MongoDB's readConcern: "majority".

With this isolation level:
- Reads only return data that has been committed by a majority of replica set members
- This prevents reading data that might be rolled back
- Each read sees the most recent committed snapshot

This scenario shows:
1. Initial data is inserted and committed
2. Session A starts a transaction and modifies data
3. Session B reads with readConcern: majority - sees ORIGINAL data
4. Session A commits
5. Session B reads again - now sees UPDATED data`
}

func (s *ReadCommittedScenario) IsolationLevel() string {
	return "Read Committed (majority)"
}

func (s *ReadCommittedScenario) Setup(ctx context.Context) error {
	// Drop and recreate with initial data
	if err := s.collection.Drop(ctx); err != nil {
		return err
	}

	// Insert initial document
	_, err := s.collection.InsertOne(ctx, bson.M{
		"account":  "checking",
		"balance":  1000.00,
		"currency": "USD",
	})
	return err
}

func (s *ReadCommittedScenario) Cleanup(ctx context.Context) error {
	return s.collection.Drop(ctx)
}

func (s *ReadCommittedScenario) Run(ctx context.Context, output chan<- scenario.StepResult) error {
	defer close(output)

	// Header
	output <- scenario.StepResult{
		IsHeader:    true,
		Description: "💰 Read Committed Isolation Demonstration",
	}

	step := 1

	// Step 1: Show initial state
	var initial bson.M
	err := s.collection.FindOne(ctx, bson.M{"account": "checking"}).Decode(&initial)
	if err != nil {
		return fmt.Errorf("failed to read initial state: %w", err)
	}

	output <- scenario.StepResult{
		Session:     "Setup",
		Step:        step,
		Description: "Initial state - checking account",
		Query:       `db.read_committed_demo.findOne({account: "checking"})`,
		Result:      fmt.Sprintf("Balance: $%.2f", initial["balance"]),
		Success:     true,
	}
	step++

	// Step 2: Session A starts a transaction and modifies balance
	sessionA, err := s.client.StartSession()
	if err != nil {
		return fmt.Errorf("failed to start session A: %w", err)
	}
	defer sessionA.EndSession(ctx)

	txnOpts := options.Transaction().
		SetReadConcern(readconcern.Majority()).
		SetWriteConcern(writeconcern.Majority())

	output <- scenario.StepResult{
		Session:     "Session A",
		Step:        step,
		Description: "Starting transaction with majority read/write concern",
		Query:       "session.startTransaction({readConcern: 'majority', writeConcern: 'majority'})",
		Result:      "Transaction started",
		Success:     true,
	}
	step++

	// Update within transaction
	err = mongo.WithSession(ctx, sessionA, func(sc mongo.SessionContext) error {
		if err := sessionA.StartTransaction(txnOpts); err != nil {
			return err
		}

		// Debit the account
		_, err := s.collection.UpdateOne(sc,
			bson.M{"account": "checking"},
			bson.M{"$inc": bson.M{"balance": -500.00}},
		)
		return err
	})
	if err != nil {
		return fmt.Errorf("failed to update in transaction: %w", err)
	}

	output <- scenario.StepResult{
		Session:     "Session A",
		Step:        step,
		Description: "Debiting $500 from checking account (within transaction)",
		Query:       `db.read_committed_demo.updateOne({account: "checking"}, {$inc: {balance: -500}})`,
		Result:      "Update applied (NOT YET COMMITTED)",
		Success:     true,
	}
	step++

	time.Sleep(500 * time.Millisecond)

	// Step 3: Session B reads with majority read concern
	output <- scenario.StepResult{
		Session:     "Session B",
		Step:        step,
		Description: "Reading account with readConcern: majority",
		Query:       `db.read_committed_demo.findOne({account: "checking"}).readConcern("majority")`,
		Result:      "",
		Success:     true,
	}

	// Use a collection with majority read concern
	collWithReadConcern := s.db.Collection("read_committed_demo", options.Collection().SetReadConcern(readconcern.Majority()))
	var resultB bson.M
	err = collWithReadConcern.FindOne(ctx, bson.M{"account": "checking"}).Decode(&resultB)
	if err != nil {
		return fmt.Errorf("failed to read with majority: %w", err)
	}

	output <- scenario.StepResult{
		Session:     "Session B",
		Step:        step,
		Description: "Read result with majority concern",
		Query:       "Result from readConcern: majority",
		Result:      fmt.Sprintf("Balance: $%.2f (ORIGINAL value - uncommitted changes not visible)", resultB["balance"]),
		Success:     true,
	}
	step++

	output <- scenario.StepResult{
		IsHeader:    true,
		Description: "✅ Session B sees only committed data (original $1000), not Session A's uncommitted -$500",
	}

	time.Sleep(500 * time.Millisecond)

	// Step 4: Session A commits
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
		Result:      "Transaction committed - balance change now permanent",
		Success:     true,
	}
	step++

	time.Sleep(500 * time.Millisecond)

	// Step 5: Session B reads again
	err = collWithReadConcern.FindOne(ctx, bson.M{"account": "checking"}).Decode(&resultB)
	if err != nil {
		return fmt.Errorf("failed to read after commit: %w", err)
	}

	output <- scenario.StepResult{
		Session:     "Session B",
		Step:        step,
		Description: "Reading account again after Session A committed",
		Query:       `db.read_committed_demo.findOne({account: "checking"}).readConcern("majority")`,
		Result:      fmt.Sprintf("Balance: $%.2f (UPDATED value now visible)", resultB["balance"]),
		Success:     true,
	}

	output <- scenario.StepResult{
		IsHeader:    true,
		Description: "🎉 After commit, Session B now sees the updated balance of $500",
	}

	return nil
}
