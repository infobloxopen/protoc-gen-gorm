package gorm

import (
	"context"

	"github.com/infobloxopen/atlas-app-toolkit/rpc/errdetails"
	"github.com/jinzhu/gorm"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sync"
)

// ctxKey is an unexported type for keys defined in this package.
// This prevents collisions with keys defined in other packages.
type ctxKey int

// txnKey is the key for `*Transaction` values in `context.Context`.
// It is unexported; clients use NewContext and FromContext
// instead of using this key directly.
var txnKey ctxKey

// NewContext returns a new Context that carries value txn.
func NewContext(parent context.Context, txn *Transaction) context.Context {
	return context.WithValue(parent, txnKey, txn)
}

// FromContext returns the *Transaction value stored in ctx, if any.
func FromContext(ctx context.Context) (txn *Transaction, ok bool) {
	txn, ok = ctx.Value(txnKey).(*Transaction)
	return
}

// Transaction serves as a wrapper around `*gorm.DB` instance.
// It works as a singleton to prevent an application of creating more than one
// transaction instance per incoming request.
type Transaction struct {
	mu      sync.Mutex
	parent  *gorm.DB
	current *gorm.DB
}

// Begin starts new transaction by calling `*gorm.DB.Begin()`
// Returns new instance of `*gorm.DB` (error can be checked by `*gorm.DB.Error`)
func (t *Transaction) Begin() *gorm.DB {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.current == nil {
		t.current = t.parent.Begin()
	}

	return t.current
}

// Rollback terminates transaction by calling `*gorm.DB.Rollback()`
// Reset current transaction and returns an error if any.
func (t *Transaction) Rollback() error {
	if t.current == nil {
		return nil
	}
	t.mu.Lock()
	defer t.mu.Unlock()

	t.current.Rollback()
	err := t.current.Error
	t.current = nil
	return err
}

// Commit finishes transaction by calling `*gorm.DB.Commit()`
// Reset current transaction and returns an error if any.
func (t *Transaction) Commit() error {
	if t.current == nil {
		return nil
	}
	t.mu.Lock()
	defer t.mu.Unlock()

	t.current.Commit()
	err := t.current.Error
	t.current = nil
	return err
}

// UnaryServerInterceptor returns grpc.UnaryServerInterceptor that manages
// a `*Transaction` instance.
// New *Transaction instance is created before grpc.UnaryHandler call.
// Client is responsible to call `txn.Begin()` to open transaction.
// If call of grpc.UnaryHandler returns with an error the transaction
// is aborted, otherwise committed.
func UnaryServerInterceptor(db *gorm.DB) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		// prepare new *Transaction instance
		txn := &Transaction{parent: db}

		defer func() {
			// simple panic handler
			if perr := recover(); perr != nil {
				// we do not try to safe the world -
				// just attempt to close our transaction
				// re-raise panic and let someone to handle it
				txn.Rollback()
				panic(perr)
			}

			var terr error
			if err != nil {
				terr = txn.Rollback()
			} else {
				if terr = txn.Commit(); terr != nil {
					err = status.Error(codes.Internal, "failed to commit transaction")
				}
			}

			if terr == nil {
				return
			}

			st := status.Convert(err)
			st, serr := st.WithDetails(errdetails.New(codes.Internal, "gorm", terr.Error()))
			// do not override error if failed to attach details
			if serr == nil {
				err = st.Err()
			}
			return
		}()

		ctx = NewContext(ctx, txn)
		resp, err = handler(ctx, req)

		return resp, err
	}
}
