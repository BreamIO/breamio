package quercy_test

import (
	"github.com/maxnordlund/breamio/briee"
	"github.com/maxnordlund/breamio/quercy"
	"testing"
)

func TestCreateDB(t *testing.T) {
	ee := briee.New()
	db, _ := quercy.New(ee, "quercy.db")
	err := db.CreateDB()
	if err != nil {
		t.Error(err)
	}
}
