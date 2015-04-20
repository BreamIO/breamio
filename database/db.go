package database

import (
	"github.com/maxnordlund/breamio/module"

	"database/sql"
	sqlite "github.com/mattn/go-sqlite3"

	"github.com/maxnordlund/breamio/eyetracker"
	"github.com/maxnordlund/breamio/moduler"
)

type QuercyFactory struct{}

func (QuercyFactory) String() string {
	return "Quercy"
}

func (QuercyFactory) New(c module.Constructor) module.Module {
	module, err := New(c)
	if err != nil {
		panic(err)
	}
	return module
}

func init() {
	moduler.Register(QuercyFactory{})
}

type DBHandler struct {
	module.SimpleModule
	*sql.DB
	insertETData      *sql.Stmt
	MethodStoreETData module.EventMethod `event:"tracker:etdata" returns:"database:errors"`
	MethodClearETData module.EventMethod `returns:"database:errors"`
}

func New(c module.Constructor) (db *DBHandler, err error) {
	raw, err := sql.Open("sqlite3", c.Parameters["source"].(string))
	db = &DBHandler{
		SimpleModule: module.NewSimpleModule("Database", c),
		DB:           raw,
	}
	if err != nil {
		return
	}
	if err = db.Ping(); err != nil {
		return
	}
	db.createTables() // Create all tables if not already there.

	return
}

//Creates all tables necessary, if they do not exist.
// We swallow all errors, because at this point, the database should be good for use.
func (db *DBHandler) createTables() error {
	return db.createETDataTable()
}

func (db *DBHandler) createETDataTable() error {
	_, err := db.Exec(`CREATE TABLE ETDATA (
		LeftX REAL,
		LeftY REAL,
		RightX REAL,
		RightY REAL,
		Timestamp INT
		);`)
	if rerr, ok := err.(sqlite.Error); ok {
		if rerr.Code == sqlite.ErrError { //SQL logical error
			//"Catch" this.
			return nil
		}
	}
	return err
}

func (db *DBHandler) StoreETData(data *eyetracker.ETData) error {
	if db.insertETData == nil {
		var err error
		db.insertETData, err = db.Prepare("INSERT INTO ETData (leftX, leftY, rightX, rightY, Timestamp) VALUES (?, ?, ?, ?, ?);")
		if err != nil {
			return err
		}
	}
	db.insertETData.Exec(data.Filtered.X(), data.Filtered.Y(), data.Filtered.X(), data.Filtered.Y(), data.Timestamp)
	return nil
}

func (db *DBHandler) ClearETData() error {
	db.Exec("DROP TABLE ETDATA;")
	return db.createETDataTable()
}
