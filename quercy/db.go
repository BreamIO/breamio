package quercy

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"

	"github.com/maxnordlund/breamio/beenleigh"
	"github.com/maxnordlund/breamio/briee"
	"github.com/maxnordlund/breamio/gorgonzola"
)

func init() {
	beenleigh.Register(&sqlRun{})
}

// Runner that starts and stops event listening for creation of new
type sqlRun struct {
}

func (s *sqlRun) Run(logic beenleigh.Logic) {

}

func (s *sqlRun) Close() error {
	return nil
}

type DBHandler struct {
	*sql.DB
	insertETData *sql.Stmt
}

func New(ee briee.EventEmitter, source string) (db *DBHandler, err error) {
	raw, err := sql.Open("sqlite3", source)
	db = &DBHandler{DB: raw}
	if err != nil {
		return
	}
	err = db.Ping()
	return
}

func (db *DBHandler) CreateDB() error {
	if _, err := db.Exec(`CREATE TABLE ETDATA (
		LeftX REAL,
		LeftY REAL,
		RightX REAL,
		RightY REAL,
		Timestamp INT
		);`); err != nil {
			return err
		}
	return nil
}

func (db *DBHandler) StoreETData(data *gorgonzola.ETData) error {
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
