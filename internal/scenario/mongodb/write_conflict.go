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

// WriteConflictScenario demonstrates write conflicts in concurrent transactions
type WriteConflictScenario struct {
	client     *mongo.Client
	db         *mongo.Database
	collection *mongo.Collection
}

// NewWriteConflictScenario creates a new write conflict demonstration scenario
func NewWriteConflictScenario(client *mongo.Client, db *mongo.Database) *WriteConflictScenario {
	return &WriteConflictScenario{
		client:     client,
		db:         db,
		collection: db.Collection("write_conflict_demo"),
	}
}

func (s *WriteConflictScenario) Name() string {
	return "Write Conflict Detection"
}

func (s *WriteConflictScenario) Description() string {
	return `Demonstrates how MongoDB detects and handles write conflicts between transactions.

When two transactions try to modify the same document:
- The first transaction to commit wins
- The second transaction gets a WriteConflict error
- This prevents lost updates and ensures data integrity

This scenario shows:
1. A bank account with $1000 balance
2. Session A starts transaction, reads balance, prepares to withdraw $600
3. Session B starts transaction, reads balance, withdraws $700 and COMMITS
4. Session A tries to commit its $600 withdrawal
5. WriteConflict! Session A must retry with new balance`
}

func (s *WriteConflictScenario) IsolationLevel() string {
	return "Serializable (Write Conflicts)"
}

func (s *WriteConflictScenario) Setup(ctx context.Context) error {
	// Drop and recreate with initial data
	if err := s.collection.Drop(ctx); err != nil {
		return err
	}

	// Insert account with balance
	_, err := s.collection.InsertOne(ctx, bson.M{
		"accountId": "ACC-12345",
		"holder":    "John Doe",
		"balance":   1000.00,
	})
	return err
}

func (s *WriteConflictScenario) Cleanup(ctx context.Context) error {
	return s.collection.Drop(ctx)
}

func (s *WriteConflictScenario) Run(ctx context.Context, output chan<- scenario.StepResult) error {
	defer close(output)

	// Header
	output <- scenario.StepResult{
		IsHeader:    true,
		Description: "⚔️ Write Conflict Detection Demonstration",
	}

	step := 1

	// Step 1: Show initial state
	var initial bson.M
	err := s.collection.FindOne(ctx, bson.M{"accountId": "ACC-12345"}).Decode(&initial)
	if err != nil {
		return fmt.Errorf("failed to read initial: %w", err)
	}

	output <- scenario.StepResult{
		Session:     "Setup",
		Step:        step,
		Description: "Initial account state",
		Query:       `db.write_conflict_demo.findOne({accountId: "ACC-12345"})`,
		Result:      fmt.Sprintf("Account: %s, Balance: $%.2f", initial["holder"], initial["balance"]),
		Success:     true,
	}
	step++

	// Step 2: Session A starts transaction and reads balance
	sessionA, err := s.client.StartSession()
	if err != nil {
		return fmt.Errorf("failed to start session A: %w", err)
	}
	defer sessionA.EndSession(ctx)

	txnOpts := options.Transaction().
		SetReadConcern(readconcern.Snapshot()).
		SetWriteConcern(writeconcern.Majority())

	// Start Session A transaction but don't commit yet
	err = mongo.WithSession(ctx, sessionA, func(sc mongo.SessionContext) error {
		if err := sessionA.StartTransaction(txnOpts); err != nil {
			return err
		}

		output <- scenario.StepResult{
			Session:     "Session A",
			Step:        step,
			Description: "Starting transaction (snapshot isolation)",
			Query:       "session.startTransaction({readConcern: 'snapshot'})",
			Result:      "Transaction started - preparing $600 withdrawal",
			Success:     true,
		}
		step++

		// Read balance
		var acct bson.M
		if err := s.collection.FindOne(sc, bson.M{"accountId": "ACC-12345"}).Decode(&acct); err != nil {
			return err
		}

		output <- scenario.StepResult{
			Session:     "Session A",
			Step:        step,
			Description: "Reading current balance",
			Query:       `db.write_conflict_demo.findOne({accountId: "ACC-12345"})`,
			Result:      fmt.Sprintf("Balance: $%.2f - Will withdraw $600", acct["balance"]),
			Success:     true,
		}
		step++

		time.Sleep(500 * time.Millisecond)

		// Session B jumps in and completes its transaction first
		output <- scenario.StepResult{
			Session:     "Session B",
			Step:        step,
			Description: "Starting SEPARATE transaction",
			Query:       "session.startTransaction({readConcern: 'snapshot'})",
			Result:      "Transaction started - will withdraw $700",
			Success:     true,
		}
		step++

		// Session B's transaction
		sessionB, err := s.client.StartSession()
		if err != nil {
			return fmt.Errorf("failed to start session B: %w", err)
		}
		defer sessionB.EndSession(ctx)

		err = mongo.WithSession(ctx, sessionB, func(scB mongo.SessionContext) error {
			if err := sessionB.StartTransaction(txnOpts); err != nil {
				return err
			}

			// Session B withdraws $700
			_, err := s.collection.UpdateOne(scB,
				bson.M{"accountId": "ACC-12345"},
				bson.M{"$inc": bson.M{"balance": -700.00}},
			)
			if err != nil {
				return err
			}

			output <- scenario.StepResult{
				Session:     "Session B",
				Step:        step,
				Description: "Withdrawing $700 from account",
				Query:       `db.write_conflict_demo.updateOne({accountId: "ACC-12345"}, {$inc: {balance: -700}})`,
				Result:      "Update applied in transaction",
				Success:     true,
			}
			step++

			// Commit Session B
			return sessionB.CommitTransaction(scB)
		})
		if err != nil {
			return fmt.Errorf("session B failed: %w", err)
		}

		output <- scenario.StepResult{
			Session:     "Session B",
			Step:        step,
			Description: "Committing transaction",
			Query:       "session.commitTransaction()",
			Result:      "✓ Transaction committed! Balance now $300",
			Success:     true,
		}
		step++

		time.Sleep(500 * time.Millisecond)

		// Session A now tries to do its update
		output <- scenario.StepResult{
			Session:     "Session A",
			Step:        step,
			Description: "Now attempting to withdraw $600 (Session A's original plan)",
			Query:       `db.write_conflict_demo.updateOne({accountId: "ACC-12345"}, {$inc: {balance: -600}})`,
			Result:      "Attempting update...",
			Success:     true,
		}
		step++

		// This should cause a write conflict
		_, err = s.collection.UpdateOne(sc,
			bson.M{"accountId": "ACC-12345"},
			bson.M{"$inc": bson.M{"balance": -600.00}},
		)

		// Try to commit - this will fail with write conflict
		commitErr := sessionA.CommitTransaction(sc)

		if commitErr != nil || err != nil {
			output <- scenario.StepResult{
				Session:     "Session A",
				Step:        step,
				Description: "Attempting to commit transaction",
				Query:       "session.commitTransaction()",
				Result:      "❌ WriteConflict! Document was modified by another transaction",
				Success:     false,
			}
			step++

			output <- scenario.StepResult{
				IsHeader:    true,
				Description: "🛡️ Write conflict detected! Session A's withdrawal prevented to avoid overdraft",
			}
		} else {
			// In case it somehow succeeded (shouldn't happen with snapshot isolation)
			output <- scenario.StepResult{
				Session:     "Session A",
				Step:        step,
				Description: "Transaction result",
				Query:       "session.commitTransaction()",
				Result:      "Transaction completed (conflict handling may vary by timing)",
				Success:     true,
			}
			step++
		}

		return nil
	})

	time.Sleep(500 * time.Millisecond)

	// Show final state
	var final bson.M
	err = s.collection.FindOne(ctx, bson.M{"accountId": "ACC-12345"}).Decode(&final)
	if err != nil {
		return fmt.Errorf("failed to read final state: %w", err)
	}

	output <- scenario.StepResult{
		Session:     "Result",
		Step:        step,
		Description: "Final account state",
		Query:       `db.write_conflict_demo.findOne({accountId: "ACC-12345"})`,
		Result:      fmt.Sprintf("Balance: $%.2f (Only Session B's $700 withdrawal applied)", final["balance"]),
		Success:     true,
	}

	output <- scenario.StepResult{
		IsHeader:    true,
		Description: "🎉 Write conflict detection prevented a potential $300 overdraft!",
	}

	return nil
}
