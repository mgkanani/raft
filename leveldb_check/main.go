package main

import (
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
//	"strconv"
	"encoding/json"
)


const (
    //DBFILE = "../leveldb2_1.db"
    DBFILE = "./leveldb2_1.db"
)

type LogItem struct{
	Index int64
	Term int
	Data interface{}
}

type DataType struct{
	Type int8 //0 to set, 1 to update,2 to delete.
	Key string
	Value interface{}
}
func main() {
Fetch()
}
func Set(){
	   	db, err := leveldb.OpenFile(DBFILE, nil)
		var log LogItem
		var temp DataType;

		temp = DataType{Type:0,Key:"abc",Value:"123"}
		log = LogItem{Index:1,Term:2,Data:temp}
		t_data, err := json.Marshal(&log)
	   	err = db.Put([]byte("1"), t_data , nil)

		temp = DataType{Type:0,Key:"pqr",Value:"456"}
		log = LogItem{Index:2,Term:2,Data:temp}
		t_data, err = json.Marshal(&log)
	   	err = db.Put([]byte("2"), t_data , nil)

		temp = DataType{Type:1,Key:"abc",Value:"789"}
		log = LogItem{Index:3,Term:3,Data:temp}
		t_data, err = json.Marshal(&log)
	   	err = db.Put([]byte("3"), t_data , nil)

	   data, err := db.Get([]byte("1"), nil)
	   fmt.Println(string(data),err)
	   err = db.Delete([]byte("key"), nil)
	   fmt.Println(data,err)
	   data, err = db.Get([]byte("key"), nil)

	   fmt.Println(data,err)
}
func Fetch(){

//	for i := 1; i < 8; i++ {
	//	DBFILE := "../leveldb2_" + strconv.Itoa(i) + ".db"
		db, err := leveldb.OpenFile(DBFILE, nil)

		iter := db.NewIterator(nil, nil)
		for iter.Next() {
			// Remember that the contents of the returned slice should not be modified, and
			// only valid until the next call to Next.
			key := iter.Key()
			value := iter.Value()
			fmt.Println(string(key), string(value))
		}
		iter.Release()
		err = iter.Error()
		fmt.Println(err)
		defer db.Close()
//	}

}
