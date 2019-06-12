package commands_test

import (
	"encoding/json"
	//"fmt"
	"os"
	"testing"

	"github.com/unswpcsoc/PCSocBot/commands"
)

/* preamble */

const INDEX = "thing"

type thing struct {
	A string `json:"name"`
	B int    `json:"age"`
	//c string `json:"unexported"`
}

func (t *thing) Index() string {
	return INDEX
}

/* actual tests */

func TestMain(m *testing.M) {
	commands.DBOpen(":memory:")

	// Create debug index, ignore errors
	tx, _ := commands.DB.Begin(true)
	tx.CreateIndex("debug", "*", func(a, b string) bool {
		return a < b
	})
	tx.Commit()

	res := m.Run()
	commands.DBClose()
	os.Exit(res)
}

/* db tests */

// TestDBGet tests DBGet
func TestDBGet(t *testing.T) {
	// Setup
	tx, err := commands.DB.Begin(true)
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

	// Set expected
	exp := thing{
		A: "first thingy",
		B: 42,
		//c: "unexported field",
	}

	// Open transaction, Set thingy in db
	tx, err = commands.DB.Begin(true)
	if err != nil {
		t.Error(err)
	}

	// Marshal struct and set it
	mar, err := json.Marshal(exp)
	if err != nil {
		t.Error(err)
	}

	// Set query
	qry := "0"
	_, _, err = tx.Set(exp.Index()+":"+qry, string(mar), nil)
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
	var got thing
	commands.DBGet(nil, qry, &got) // empty Storer
	commands.DBGet(&exp, "", &got) // empty query
	commands.DBGet(&exp, "", nil)  // empty ptr
	commands.DBGet(nil, "", nil)   // empty everything

	// Good query
	err = commands.DBGet(&exp, qry, &got)
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
	if got != exp {
		t.Errorf("DBGet(%s) = %#v; want %#v", qry, got, exp)
	}
}

// TestDBSet tests DBSet
func TestDBSet(t *testing.T) {
	// Setup
	tx, err := commands.DB.Begin(true)
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

	exp := thing{
		A: "first thingy",
		B: 42,
		//c: "unexported field",
	}

	ind := "1"
	// Bad sets, make sure nothing panics
	commands.DBSet(nil, ind) // empty storer
	commands.DBSet(&exp, "") // empty key
	commands.DBSet(nil, "")  // empty everything

	// Set exp in db
	_, _, err = commands.DBSet(&exp, ind)
	if err != nil {
		t.Error(err)
	}

	// Open RO transaction
	tx, err = commands.DB.Begin(false)
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
	var got thing
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
	if got != exp {
		t.Errorf("DBSet(%[1]s, %#[3]v) set {%[2]s: %#[4]v}; want {%[2]s: %#[5]v}", ind, qry, exp, got, exp)
	}
}
