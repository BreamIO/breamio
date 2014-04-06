package quercy

import (
	bl "github.com/maxnordlund/breamio/beenleigh"
	"github.com/maxnordlund/breamio/briee"
	g "github.com/maxnordlund/breamio/gorgonzola"
	//"github.com/maxnordlund/breamio/quercy"

	"database/sql"
	"testing"
	"time"
)

func setup(t *testing.T) (briee.EventEmitter, *DBHandler) {
	ee := briee.New()
	dbh, _ := New(ee, "quercy_test.db")
	return ee, dbh
}

func teardown(dbh *DBHandler, t *testing.T) {
	defer dbh.Close()
	//t.Log("Teardown.")
	if _, err := dbh.Exec("DROP TABLE ETDATA;"); err != nil {
		t.Error("teardown:", err)
	}
}

//Currently only tests for panics.
//Checking listening is sort of performed in other tests.
func TestRunner(t *testing.T) {
	runner := &sqlRun{make(chan struct{})}
	logic := bl.New(briee.New)
	go runner.Run(logic)
	defer runner.Close()
	time.Sleep(time.Second) //Wait for subscriptions to be setup.
	errorCh := logic.RootEmitter().Subscribe("storage:error", string("")).(<-chan string)

	logic.RootEmitter().Dispatch("new:storage", bl.Spec{1, "quercy_test.db"})
	select {
	case err := <-errorCh:
		t.Error(err)
	case <-time.After(time.Second):
	}
	logic.RootEmitter().Dispatch("storage:shutdown", struct{}{})
}

// Method should create the tables needed by the rest of the applications.
func TestCreateTables(t *testing.T) {
	_, dbh := setup(t)
	defer teardown(dbh, t)
	dbh.createTables()

	db, _ := sql.Open("sqlite3", "quercy_test.db")
	defer db.Close()
	_, err := db.Query("Select * from ETDATA;")
	if err != nil {
		t.Error(err)
	}
}

func TestBadSource(t *testing.T) {
	ee := briee.New()
	dbh, err := New(ee, "Ö:\\bad.db")
	defer dbh.Close()
	if err == nil {
		t.Error("Ö:\\bad.db should not be a valid source.")
	}
}

func TestCreateOnUnconnectedHandler(t *testing.T) {
	_, dbh := setup(t)
	//defer teardown(dbh, t)
	dbh.Close()
	err := dbh.createTables()
	if err == nil {
		t.Error("Should return a error.")
	}

}

// Method should verify the integrity of the database,
// and decide if it needs to be initialized
func TestVerify(t *testing.T) {
	t.Skip("NOT IMPLEMENTED")
}

func TestChangeSource(t *testing.T) {
	t.Skip("NOT IMPLEMENTED")
}

func TestClearETData(t *testing.T) {
	_, dbh := setup(t)
	defer teardown(dbh, t)
	dbh.createETDataTable()
	dbh.StoreETData(&g.ETData{g.Point2D{0.1, 0.1}, time.Now()})
	if err := dbh.ClearETData(); err != nil {
		t.Error(err)
	}
	results, err := dbh.Query("SELECT LeftX, LeftY FROM ETDATA")
	if err != nil {
		t.Error(err)
	}
	if results.Next() {
		t.Error("No data should remain.")
	}
}

func TestClearWithNoTable(t *testing.T) {
	_, dbh := setup(t)
	defer teardown(dbh, t)
	if err := dbh.ClearETData(); err != nil {
		t.Error(err)
	}
	_, err := dbh.Query("SELECT LeftX, LeftY FROM ETDATA")
	if err != nil {
		t.Error(err)
	}
}

func TestStoreETData(t *testing.T) {
	_, dbh := setup(t)
	defer teardown(dbh, t)
	dbh.createETDataTable()
	for i := float64(0); i < 11; i++ {
		err := dbh.StoreETData(&g.ETData{g.Point2D{0.123 + 0.1*i, 0.456 + 0.01*i}, time.Now()})
		if err != nil {
			t.Fatal(err)
		}
	}
	results, _ := dbh.Query("Select LeftX,LeftY from ETDATA;")
	defer results.Close()
	ok := results.Next()
	if !ok {
		t.Fatal("No records stored.")
	}

	var x, y float64
	results.Scan(&x, &y)
	if x != 0.123 || y != 0.456 {
		t.Fatalf("Wrong data stored. Expected (%f, %f), found (%f, %f).", 0.123, 0.456, x, y)
	}
}

func TestStoreETDataInClosedHandler(t *testing.T) {
	_, dbh := setup(t)
	defer dbh.Exec("DROP TABLE ETDATA;")
	dbh.Close()
	dbh.createETDataTable()
	err := dbh.StoreETData(&g.ETData{g.Point2D{0.333, 0.333}, time.Now()})
	if err == nil {
		t.Fatal(err)
	}
}

func TestStoreETDataFromEvent(t *testing.T) {
	ee, dbh := setup(t)
	defer teardown(dbh, t)
	dbh.createETDataTable()

	errCh := ee.Subscribe("storage:error", string("")).(<-chan string)
	//defer ee.Unsubscribe("storage:error", errCh)
	for i := float64(0); i < 5; i++ {
		ee.Dispatch("tracker:etdata", &g.ETData{g.Point2D{0.456 + 0.1*i, 0.123 + 0.01*i}, time.Now()})
	}

	//Check
	select {
	case err := <-errCh:
		t.Fatal(err)
	case <-time.After(time.Second): //Timeout. No errors occured.
	}
	results, _ := dbh.Query("Select LeftX,LeftY from ETDATA;")
	defer results.Close()
	ok := results.Next()
	if !ok {
		t.Fatal("No records stored.")
	}

	var x, y float64
	results.Scan(&x, &y)
	if x != 0.456 || y != 0.123 {
		t.Fatalf("Wrong data stored. Expected (%f, %f), found (%f, %f).", 0.123, 0.456, x, y)
	}
}

func TestEventClose(t *testing.T) {
	ee, dbh := setup(t)
	ee.Dispatch("storage:shutdown", struct{}{})
	select {
	case <-dbh.closer:
	case <-time.After(time.Second):
		t.Error("Timed out on closer listening.")
	}
}
