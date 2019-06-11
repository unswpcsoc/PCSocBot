package commands

import (
	"encoding/json"
	//"fmt"
	"os"
	"testing"
)

/* preamble */

const (
	INDEX = "thing"
)

type thing struct {
	A string `json:"name"`
	B int    `json:"age"`
	//c string `json:"unexported"`
}

func (t *thing) Index() string {
	return INDEX
}

func (t *thing) Unmarshal(mar string) (Storer, error) {
	var res thing
	err := json.Unmarshal([]byte(mar), &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

/* actual tests */

func TestMain(m *testing.M) {
	DBOpen(":memory:")

	// Create debug index, ignore errors
	tx, _ := db.Begin(true)
	tx.CreateIndex("debug", "*", func(a, b string) bool {
		return a < b
	})
	tx.Commit()

	res := m.Run()
	DBClose()
	os.Exit(res)
}

/* db tests */

// TestDBGet tests DBGet
func TestDBGet(t *testing.T) {
	// Setup
	tx, err := db.Begin(true)
	if err != nil {
		t.Error(err)
	}

	err = tx.DeleteAll()
	if err != nil {
		t.Error(err)
	}

	err = tx.Commit()
	if err != nil {
		t.Error(err)
	}

	thingy := &thing{
		A: "first thingy",
		B: 42,
		//c: "unexported field",
	}

	// Open transaction, Set thingy in db
	tx, err = db.Begin(true)
	if err != nil {
		t.Error(err)
	}

	// Set expected
	exp := thingy

	// Marshal struct and set it
	mar, err := json.Marshal(*thingy)
	if err != nil {
		t.Error(err)
	}

	// Set query
	qry := "0"
	_, _, err = tx.Set(thingy.Index()+":"+qry, string(mar), nil)
	if err != nil {
		t.Error(err)
	}

	// Write to db
	err = tx.Commit()
	if err != nil {
		t.Error(err)
	}
	tx = nil

	// Bad queries, make sure nothing panics
	DBGet(nil, qry)   // empty storer
	DBGet(thingy, "") // empty query
	DBGet(nil, "")    // empty everything

	// Good query
	got, err := DBGet(thingy, qry)
	if err != nil {
		t.Error(err)
	}

	/* Get whole db, if you wish
	tx, err = db.Begin(false)
	if err != nil {
		t.Error(err)
	}
	defer tx.Rollback()
	buf := ""
	tx.Ascend("debug", func(key, value string) bool {
		buf += "\t" + key + " : " + value + "\n"
		return true
	})
	fmt.Printf("TestDBGet: DB HAS {\n%s}\n", buf)
	*/

	// Compare got and exp
	// NB: 	pointer receiver methods are added to the type's pointer's method table,
	// 		type asserting as the pointer is required here because of that
	if *got.(*thing) != *exp {
		t.Errorf("DBGet(%s) = %#v; want %#v", qry, got, exp)
	}
}

// TestDBSet tests DBSet
func TestDBSet(t *testing.T) {
	// Setup
	tx, err := db.Begin(true)
	if err != nil {
		t.Error(err)
	}

	err = tx.DeleteAll()
	if err != nil {
		t.Error(err)
	}

	err = tx.Commit()
	if err != nil {
		t.Error(err)
	}

	thingy := &thing{
		A: "first thingy",
		B: 42,
		//c: "unexported field",
	}

	// Set expected
	exp := thingy

	ind := "1"
	// Bad sets, make sure nothing panics
	DBSet(nil, ind)   // empty storer
	DBSet(thingy, "") // empty key
	DBSet(nil, "")    // empty everything

	// Set thingy in db
	_, _, err = DBSet(thingy, ind)
	if err != nil {
		t.Error(err)
	}

	// Open RO transaction
	tx, err = db.Begin(false)
	if err != nil {
		t.Error(err)
	}
	defer tx.Rollback()

	// Query db
	qry := INDEX + ":" + ind
	res, err := tx.Get(qry)
	if err != nil {
		t.Error(err)
	}

	// Unmarshal into struct
	var got *thing
	err = json.Unmarshal([]byte(res), &got)
	if err != nil {
		t.Error(err)
	}

	/* Get whole db, if you wish
	buf := ""
	tx.Ascend("debug", func(key, value string) bool {
		buf += "\t" + key + " : " + value + "\n"
		return true
	})
	fmt.Printf("TestDBSet: DB HAS {\n%s}\n", buf)
	*/

	// Compare got and exp
	if *got != *exp {
		t.Errorf("DBSet(%[1]s, %#[3]v) set {%[2]s: %#[4]v}; want {%[2]s: %#[5]v}", ind, qry, thingy, got, exp)
	}
}
