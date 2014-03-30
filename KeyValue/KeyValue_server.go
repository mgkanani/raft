package main

import (
	"encoding/json"
	//	"flag"
	"fmt"
	Raft "github.com/mgkanani/raft"
	"github.com/syndtr/goleveldb/leveldb"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"

//	"reflect"
)

const (
	SET    = 0
	GET    = 3
	UPDATE = 1
	DELETE = 2
)

var Map map[string]string
var Pending map[int64]bool

type MsgStruct struct {
	Key  int64
	Data interface{}
}

/*
type Raft.DataType struct {
	Type  int8 //0 to set, 1 to update,2 to delete.
	Key   string
	Value interface{}
}
*/
var rafttype *Raft.RaftType

var myid = 0

//var mut sync.Mutex
var mut sync.RWMutex

func main() {
	if len(os.Args) != 3 {
		fmt.Println("format:-KeyValue ipaddr:port pid")
		return
	}
	//ch := make(chan bool)
	Map = make(map[string]string)
	Pending = make(map[int64]bool)

	socket := os.Args[1]
	addr_port, err := net.ResolveTCPAddr("tcp", socket)
	if err != nil {
		fmt.Println("Address format is ipaddr:port_num")
		return
	}
	ln, err := net.ListenTCP("tcp", addr_port)
	if err != nil {
		fmt.Println("use different port number, given port already in use.\nin", err.Error())
		return
	}

	fmt.Println("Listening at addr:port", addr_port)
	defer ln.Close()

	/*	flagid := flag.Int("pid", 1, "flag type is integer")
		flag.Parse()
		fmt.Println(*flagid)
		myid := *flagid
		//wg := new(sync.WaitGroup)
	*/
	var valid bool
	//ch := make(chan int)
	//valid := InitServer(myid, "./config.json",true , ch)
	myid, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	flagid := &myid
	fmt.Println(os.Args, os.Args[2], myid)
	ConstructKeyValue(flagid)

	valid, rafttype = Raft.InitServer(myid, "./config.json", true)

	fmt.Println("valid:-", valid)
	if valid {
		var is_leader int
		for {
			conn, err := ln.Accept()
			rafttype.Leader(&is_leader)
			fmt.Println("is_leader:-", is_leader)
			if myid == is_leader {
				if err != nil {
					fmt.Print(err)
					return
				}
				go handle_client(conn)
			} else {
				_, _ = conn.Write([]byte("Please contact Server -" + strconv.Itoa(rafttype.Vote())))
				conn.Close()
			}
		}
	} else {
		log.Println("error generated in starting server according to configuration file")
		os.Exit(1)
	}
}

func Get_Val(key string) string {
	return Map[key]
}

func Set_Val(key string, val string) {
	Map[key] = val
	return
}

func Update_Val(key string, val string) {
	Map[key] = val
	return
}

func Delete(key string) {
	delete(Map, key)
	return
}

func Rename(key1 string, key2 string) bool {
	ok := true
	_, ok = Map[key2] //first check key2 must not consist any-value.
	if ok {
		return !ok
	}
	Map[key2] = Map[key1]
	delete(Map, key1)
	return !ok
}

func handleReq(msg Raft.DataType) {
	if true {
		index := rafttype.GetIndex(msg)
		litem := &Raft.LogItem{Index: index, Term: int64(rafttype.Term()), Data: msg}
		fmt.Println(litem)
		rafttype.Inbox() <- litem
		for {
			data := <-rafttype.Outbox()
			//fmt.Println(reflect.TypeOf(data),*(data.(*int64)),index)
			if *(data.(*int64)) == index {
				break
			}
		}
		//Pending[index] = false
	}
}

func ConstructKeyValue(pid *int) {
	DBFILE := "./leveldb2"
	DBFILE += "_" + strconv.Itoa(*pid) + ".db"
	log.Println(*pid, DBFILE)
	db, err := leveldb.OpenFile(DBFILE, nil)
	if err != nil {
		log.Println("err in opening file for leveldb:-", DBFILE, "error is:-", err)
	}
	iter := db.NewIterator(nil, nil)

	var logitem Raft.LogItem
	var content Raft.DataType

	for iter.Next() {
		CommitIndex, err := strconv.ParseInt(string(iter.Key()), 10, 64)
		fmt.Println(CommitIndex, err)
		err = json.Unmarshal(iter.Value(), &logitem) //decode message into Envelope object.
		fmt.Println("response of conversion from log to log item", logitem)
		//  now set/get/delete key-value in Map based on Log
		if logitem.Data != nil {
			testing := logitem.Data.(map[string]interface{})
			/*
				for key, value := range testing{
					fmt.Println(reflect.TypeOf(value),key,value)
				}
			*/
			fmt.Println(testing["Type"])
			content = Raft.DataType{Type: int8(int(testing["Type"].(float64))), Key: testing["Key"].(string), Value: testing["Value"]}
			fmt.Println(content)
			switch content.Type {
			case 0, 1: //set value
				Map[content.Key] = content.Value.(string)
				break
			case 2: //delete value
				delete(Map, content.Key)
				break
			}
		}
	}
	iter.Release()
	log.Println("Error if:-", iter.Error(), "Map is:-", Map)
	defer db.Close()

}

func handle_client(c net.Conn) {
	defer c.Close()
	/*
	   	var is_leader int;
	   	rafttype.Leader(&is_leader)
	           if myid == is_leader {
	*/
	for {
		msg := make([]byte, 1000)

		n, err := c.Read(msg)
		if err == io.EOF {
			fmt.Printf("SERVER: received EOF (%d bytes ignored)\n", n)
			return
		} else if err != nil {
			fmt.Printf("ERROR: read\n")
			fmt.Print(err)
			return
		}
		fmt.Printf("SERVER: received %v bytes\n", n)

		//to_send :=string(msg);
		str := strings.Fields(string(msg))

		fmt.Printf("msg=%s\n", msg)
		//fmt.Println(str[0]=="get",str[1])
		action := str[0]
		key := str[1]
		switch {

		case action == "set":
			val := str[2]
			msg := Raft.DataType{Type: SET, Key: key, Value: val}
			handleReq(msg)
			mut.Lock()
			//Map[key] = val
			Set_Val(key, val)
			mut.Unlock()
			fmt.Printf("set Map[%s]=%s", key, val)

		case action == "update":
			val := str[2]
			msg := Raft.DataType{Type: UPDATE, Key: key, Value: val}
			handleReq(msg)
			mut.Lock()
			//Map[key] = val
			Update_Val(key, val)
			mut.Unlock()
			fmt.Printf("update Map[%s]=%s", str[1], str[2])

		case action == "get":
			//fmt.Printf("s Map[%s]=%s e\n", str[1], Map[key])
			//fmt.Printf("s Map[%s]=%s e\n", str[1], Map["abc"])
			//temp:=string(strings.TrimSpace(str[1]));
			mut.RLock()
			//temp := []byte(Map[key])
			temp := []byte(Get_Val(key))
			mut.RUnlock()
			//fmt.Println(temp)
			//fmt.Printf("get Map[%s]=%s", key, temp)
			n, err = c.Write(temp)
			if n == 0 {
				n, err = c.Write([]byte("\n"))
			}

			if err != nil {
				fmt.Printf("ERROR: write\n")
				fmt.Print(err)
				return
			}
			fmt.Printf("SERVER: sent %v bytes\n", n)

		case action == "delete":
			msg := Raft.DataType{Type: DELETE, Key: key}
			handleReq(msg)
			mut.Lock()
			Delete(key)
			mut.Unlock()
			fmt.Printf("delete Map[%s]", key)

		case action == "rename":
			keyfrom := str[1]
			keyto := str[2]
			mut.Lock()
			Rename(keyfrom, keyto)
			mut.Unlock()
			fmt.Printf("Rename from Map[%s] to Map[%s]", keyfrom, keyto)

		}
	}
	/*
		}else{
			_,_ = c.Write([]byte("Please contact Server - "+strconv.Itoa(rafttype.Vote())))

		}
	*/
}
