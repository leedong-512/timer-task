package conn

import (
	"log"

	"github.com/tidwall/buntdb"
)

var BuntDb *buntdb.DB

func init() {
	BuntDb = NewBuntdbClient()
}

func NewBuntdbClient() *buntdb.DB {
	buntDb, err := buntdb.Open("../timer-task-manager/data/data.db")
	if err != nil {
		log.Fatal(err)
	}
	return buntDb
}
