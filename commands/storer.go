package commands

import (
	"encoding/json"
	"errors"
	"reflect"
	"sync"

	"github.com/tidwall/buntdb"
)

var (
	// DB is the database
	DB *buntdb.DB

	// ErrDBNotOpen means db wasn't opened when trying to use it
	ErrDBNotOpen = errors.New("db not open, use DBOpen()")
	// ErrDBClosed means db was closed already
	ErrDBClosed = errors.New("db already closes")
	// ErrDBValueNil means you tried to set a nil value into the db
	ErrDBValueNil = errors.New("cannot set nil value")
	// ErrDBKeyEmpty means you tried to set a value without a key
	ErrDBKeyEmpty = errors.New("cannot set value with empty key")
	// ErrDBNotPtr means you didn't give a pointer
	ErrDBNotPtr = errors.New("want pointer arg, got something else")
	// ErrDBNotFound means there is no db
	ErrDBNotFound = buntdb.ErrNotFound

	// ErrStorerNil means you have made bad life decisions
	ErrStorerNil = errors.New("storer method received nil")

	lock = &sync.Mutex{}
	once = &sync.Once{}
)

/* db stuff */

// Storer is the interface for structs that will be stored into the db
//
// You MUST export ALL fields in a Storer, otherwise the JSON Marshaller will freak out
// there are workarounds, but they require more effort than we need.
// Read https://stackoverflow.com/a/49372417 if you're interested
type Storer interface {
	Index() string // Determines db index
}

// DBOpen opens the db at the given path
func DBOpen(path string) error {
	var err error
	DB, err = buntdb.Open(path)
	DB.Shrink()
	return err
}

// DBClose closes the db
func DBClose() error {
	if DB == nil {
		return ErrDBClosed
	}
	err := DB.Close()
	if err != nil {
		return err
	}
	DB = nil
	return nil
}

// DBSet is a Storer method that sets the given Storer in the db at the key.
func DBSet(s Storer, key string) (previous string, replaced bool, err error) {
	// Assert db open so we can rollback transactions on later errors
	if DB == nil {
		return "", false, ErrDBNotOpen
	}
	if s == nil {
		return "", false, ErrStorerNil
	}
	if len(key) == 0 {
		return "", false, ErrDBKeyEmpty
	}

	// Begin RW transaction
	tx, err := DB.Begin(true)
	if err != nil {
		tx.Rollback()
		return "", false, err
	}

	// Marshal storer
	mar, err := json.Marshal(s)
	if err != nil {
		tx.Rollback()
		return "", false, err
	}

	// Set marshalled key/value pair
	pre, rep, err := tx.Set(s.Index()+":"+key, string(mar), nil)
	if err != nil {
		tx.Rollback()
		return "", false, err
	}

	// Commit changes
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return "", false, err
	}

	return pre, rep, nil
}

// DBGet gets the Storer at the given key and puts it into got. Ignores expiry.
//
// If got is not a pointer, DBGet will throw ErrDBNotPtr
func DBGet(s Storer, key string, got Storer) error {
	if DB == nil {
		return ErrDBNotOpen
	}
	if s == nil || got == nil {
		return ErrStorerNil
	}
	if reflect.TypeOf(got).Kind() != reflect.Ptr {
		return ErrDBNotPtr
	}

	// Open RO Transaction, defer rollback
	tx, err := DB.Begin(false)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get Storer
	res, err := tx.Get(s.Index()+":"+key, true)
	if err != nil {
		return err
	}

	// Unmarshal Storer
	return json.Unmarshal([]byte(res), got)
}

// DBLock locks the db
func DBLock() { lock.Lock() }

// DBUnlock unlocks the db
func DBUnlock() { lock.Unlock() }

// DBNewOnce refreshes the once primitive
func DBNewOnce() { once = &sync.Once{} }

// DBOnce uses the once primitive
func DBOnce(do func()) { once.Do(do) }
