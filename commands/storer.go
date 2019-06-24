package commands

import (
	"encoding/json"
	"errors"
	"reflect"

	"github.com/tidwall/buntdb"
)

var (
	/* db */
	DB *buntdb.DB

	/* db errors */
	ErrDBClosed   = errors.New("db not open, use DBOpen()")
	ErrDBOpen     = errors.New("db already open")
	ErrDBValueNil = errors.New("cannot set nil value")
	ErrDBKeyEmpty = errors.New("cannot set value with empty key")
	ErrDBNotPtr   = errors.New("want pointer arg, got something else")
	ErrDBNotFound = buntdb.ErrNotFound

	/* storer errors */
	ErrStorerNil = errors.New("storer method received nil")
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

// Open opens the db at the given path
func DBOpen(path string) error {
	var err error
	DB, err = buntdb.Open(path)
	return err
}

// Close closes the db
func DBClose() error {
	if DB == nil {
		return ErrDBOpen
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
		return "", false, ErrDBClosed
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
		return ErrDBClosed
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
