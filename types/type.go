package types

import (
	"time"
)

// MiftahValue represents possible types of values that can be stored in the database
type MiftahValue interface{}

// IMiftahDB defines the public methods for interacting with MiftahDB
type IMiftahDB interface {
	Get(key string) (MiftahValue, error)
	Set(key string, value MiftahValue, expiresAt *time.Time) error
	GetExpire(key string) (*time.Time, error)
	SetExpire(key string, expiresAt time.Time) error
	Exists(key string) bool
	Delete(key string) error
	Rename(oldKey, newKey string) error
	Keys(pattern string) ([]string, error)
	Pagination(limit, page int, pattern string) ([]string, error)
	Count(pattern string) (int, error)
	CountExpired(pattern string) (int, error)
	MultiGet(keys []string) (map[string]MiftahValue, error)
	MultiSet(entries []Entry) error
	MultiDelete(keys []string) error
	Cleanup() error
	Vacuum() error
	Close() error
	Flush() error
	Backup(path string) error
	Restore(path string) error
	Execute(sql string, params ...interface{}) (interface{}, error)
}

// Entry represents a key-value pair with an optional expiration time
type Entry struct {
	Key       string      // The key of the entry
	Value     MiftahValue // The value associated with the key
	ExpiresAt *time.Time  // Expiration time (optional)
}

// MiftahDBItem represents an item stored in the database
type MiftahDBItem struct {
	Value     []byte // The stored value as a byte array
	ExpiresAt int64  // The expiration timestamp in milliseconds, or null if no expiration
}
