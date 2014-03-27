package main

import (
    "github.com/syndtr/goleveldb/leveldb"
    "fmt"
    "strconv"
)
/*
const (
    DBFILE = "../leveldb2_1.db"
)
*/
func main() {


/*
err = db.Put([]byte("key"), []byte("value"), nil)

data, err := db.Get([]byte("key"), nil)
fmt.Println(data,err)
err = db.Delete([]byte("key"), nil)
fmt.Println(data,err)
data, err = db.Get([]byte("key"), nil)

fmt.Println(data,err)
*/



for i:=1;i<8;i++{
	DBFILE := "../leveldb2_"+strconv.Itoa(i)+".db"
	db, err := leveldb.OpenFile(DBFILE, nil)

iter := db.NewIterator(nil, nil)
for iter.Next() {
    // Remember that the contents of the returned slice should not be modified, and
    // only valid until the next call to Next.
    key := iter.Key()
    value := iter.Value()
    fmt.Println(string(key),string(value))
}
iter.Release()
err = iter.Error()
fmt.Println(err);
defer db.Close()
}

}

