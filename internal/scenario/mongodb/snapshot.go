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
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

// SnapshotIsolationScenario demonstrates snapshot isolation in MongoDB
type SnapshotIsolationScenario struct {
	client     *mongo.Client
	db         *mongo.Database
	collection *mongo.Collection
}

// NewSnapshotIsolationScenario creates a new snapshot isolation demonstration scenario
func NewSnapshotIsolationScenario(client *mongo.Client, db *mongo.Database) *SnapshotIsolationScenario {
	return &SnapshotIsolationScenario{
		client:     client,
		db:         db,
		collection: db.Collection("snapshot_demo"),
	}
}

func (s *SnapshotIsolationScenario) Name() string {
	return "Snapshot Isolation"
}

func (s *SnapshotIsolationScenario) Description() string {
	return `Demonstrates Snapshot Isolation using MongoDB's readConcern: "snapshot".

This is the strongest isolation level in MongoDB. With snapshot isolation:
- All reads in a transaction see data from the same point in time
- The snapshot is taken at the START of the transaction
- Changes committed by other transactions AFTER your transaction starts are INVISIBLE

This scenario shows:
1. Initial inventory with 3 products
2. Session A starts a transaction with snapshot isolation
3. Session A reads inventory - sees 3 products
4. Session B adds a new product and COMMITS immediately
5. Session A reads again - STILL sees only 3 products (snapshot!)
6. After Session A ends, new product becomes visible`
}

func (s *SnapshotIsolationScenario) IsolationLevel() string {
	return "Snapshot (Repeatable Read)"
}

func (s *SnapshotIsolationScenario) Setup(ctx context.Context) error {
	// Drop and recreate with initial data
	if err := s.collection.Drop(ctx); err != nil {
		return err
	}

	// Insert initial products
	_, err := s.collection.InsertMany(ctx, []interface{}{
		bson.M{"sku": "WIDGET-001", "name": "Blue Widget", "quantity": 100},
		bson.M{"sku": "WIDGET-002", "name": "Red Widget", "quantity": 50},
		bson.M{"sku": "GADGET-001", "name": "Super Gadget", "quantity": 25},
	})
	return err
}

func (s *SnapshotIsolationScenario) Cleanup(ctx context.Context) error {
	return s.collection.Drop(ctx)
}

func (s *SnapshotIsolationScenario) Run(ctx context.Context, output chan<- scenario.StepResult) error {
	defer close(output)

	// Header
	output <- scenario.StepResult{
		IsHeader:    true,
		Description: "📸 Snapshot Isolation Demonstration",
	}

	step := 1

	// Step 1: Show initial state
	count, err := s.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("failed to count initial: %w", err)
	}

	output <- scenario.StepResult{
		Session:     "Setup",
		Step:        step,
		Description: "Initial inventory state",
		Query:       "db.snapshot_demo.countDocuments({})",
		Result:      fmt.Sprintf("Product count: %d (Blue Widget, Red Widget, Super Gadget)", count),
		Success:     true,
	}
	step++

	// Step 2: Session A starts transaction with snapshot isolation
	sessionA, err := s.client.StartSession()
	if err != nil {
		return fmt.Errorf("failed to start session A: %w", err)
	}
	defer sessionA.EndSession(ctx)

	txnOpts := options.Transaction().
		SetReadConcern(readconcern.Snapshot()).
		SetWriteConcern(writeconcern.Majority())

	var snapshotCount int64

	err = mongo.WithSession(ctx, sessionA, func(sc mongo.SessionContext) error {
		if err := sessionA.StartTransaction(txnOpts); err != nil {
			return err
		}

		output <- scenario.StepResult{
			Session:     "Session A",
			Step:        step,
			Description: "Starting transaction with SNAPSHOT isolation",
			Query:       "session.startTransaction({readConcern: 'snapshot'})",
			Result:      "Transaction started - snapshot of database taken NOW",
			Success:     true,
		}
		step++

		// Read count within transaction
		snapshotCount, err = s.collection.CountDocuments(sc, bson.M{})
		if err != nil {
			return err
		}

		output <- scenario.StepResult{
			Session:     "Session A",
			Step:        step,
			Description: "Reading product count within snapshot transaction",
			Query:       "db.snapshot_demo.countDocuments({})",
			Result:      fmt.Sprintf("Product count: %d", snapshotCount),
			Success:     true,
		}
		step++

		time.Sleep(500 * time.Millisecond)

		// Session B (outside transaction) inserts a new product
		output <- scenario.StepResult{
			Session:     "Session B",
			Step:        step,
			Description: "Inserting NEW product (outside of Session A's transaction)",
			Query:       `db.snapshot_demo.insertOne({sku: "GADGET-002", name: "Ultra Gadget", quantity: 10})`,
			Result:      "",
			Success:     true,
		}

		// Insert using a separate context (not in transaction)
		_, err = s.collection.InsertOne(ctx, bson.M{
			"sku":      "GADGET-002",
			"name":     "Ultra Gadget",
			"quantity": 10,
		})
		if err != nil {
			return fmt.Errorf("session B insert failed: %w", err)
		}

		output <- scenario.StepResult{
			Session:     "Session B",
			Step:        step,
			Description: "New product inserted and COMMITTED immediately",
			Query:       "Insert completed with default write concern",
			Result:      "New product 'Ultra Gadget' is now in the database",
			Success:     true,
		}
		step++

		time.Sleep(500 * time.Millisecond)

		// Verify Session B can see it (outside transaction)
		totalCount, err := s.collection.CountDocuments(ctx, bson.M{})
		if err != nil {
			return err
		}

		output <- scenario.StepResult{
			Session:     "Session B",
			Step:        step,
			Description: "Session B verifies new product exists",
			Query:       "db.snapshot_demo.countDocuments({})",
			Result:      fmt.Sprintf("Product count: %d (Session B sees 4 products)", totalCount),
			Success:     true,
		}
		step++

		time.Sleep(500 * time.Millisecond)

		// Session A reads again - should STILL see old snapshot
		snapshotCount, err = s.collection.CountDocuments(sc, bson.M{})
		if err != nil {
			return err
		}

		output <- scenario.StepResult{
			Session:     "Session A",
			Step:        step,
			Description: "Session A reads product count AGAIN (still in same transaction)",
			Query:       "db.snapshot_demo.countDocuments({})",
			Result:      fmt.Sprintf("Product count: %d (SNAPSHOT - doesn't see new product!)", snapshotCount),
			Success:     true,
		}
		step++

		output <- scenario.StepResult{
			IsHeader:    true,
			Description: "✅ Snapshot isolation in action! Session A still sees 3 products, even though Session B committed 4th",
		}

		// Commit Session A's transaction
		return sessionA.CommitTransaction(sc)
	})
	if err != nil {
		return fmt.Errorf("session A transaction failed: %w", err)
	}

	output <- scenario.StepResult{
		Session:     "Session A",
		Step:        step,
		Description: "Committing Session A's transaction",
		Query:       "session.commitTransaction()",
		Result:      "Transaction committed - snapshot released",
		Success:     true,
	}
	step++

	time.Sleep(500 * time.Millisecond)

	// Now read outside any transaction
	finalCount, err := s.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("failed to count final: %w", err)
	}

	output <- scenario.StepResult{
		Session:     "Session A",
		Step:        step,
		Description: "Session A reads after transaction ends",
		Query:       "db.snapshot_demo.countDocuments({})",
		Result:      fmt.Sprintf("Product count: %d (Now sees all products including Ultra Gadget)", finalCount),
		Success:     true,
	}

	output <- scenario.StepResult{
		IsHeader:    true,
		Description: "🎉 Snapshot isolation provides a consistent view throughout the entire transaction",
	}

	return nil
}
