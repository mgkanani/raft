package main

import (
	"encoding/json"
	"flag"
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
	//"reflect"
)

var Map map[string]string

type MsgStruct struct {
	Key  int64
	Data interface{}
}

type DataType struct {
	Type  int8 //0 to set, 1 to update,2 to delete.
	Key   string
	Value interface{}
}

var rafttype *Raft.RaftType
var myid = 0

//var mut sync.Mutex
var mut sync.RWMutex

func main() {
	if len(os.Args) != 4 {
		fmt.Println("format:-KeyValue ipaddr:port -pid n")
		return
	}
	//ch := make(chan bool)
	Map = make(map[string]string)

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

	flagid := flag.Int("pid", 1, "flag type is integer")
	flag.Parse()
	myid = *flagid
	//wg := new(sync.WaitGroup)
	var valid bool
	//ch := make(chan int)
	//valid := InitServer(myid, "./config.json",true , ch)
	constructKeyValue(myid)

	valid, rafttype = Raft.InitServer(myid, "./config.json", true)

	if valid {
		for {
			conn, err := ln.Accept()
			if err != nil {
				fmt.Print(err)
				return
			}
			go handle_client(conn)
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

func handleReq(msg []byte) {
	if true {
		litem := &Raft.LogItem{Index: rafttype.GetIndex(), Term: int64(rafttype.Term()), Data: msg}
		fmt.Println(litem)
		//rafttype.Inbox() <- litem
		/*
			var is_leader int
			rafttype.Leader(&is_leader)
			if myid == is_leader {
				litem := &Raft.LogItem{Index: 1, Data: "hello"}
				fmt.Println("ch1", rafttype)
				in := rafttype.Inbox()
				fmt.Println("ch2", in)
				in <- litem
			}
		*/
	}
}

func constructKeyValue(pid int) {
	/*
	   //Assumptions:-
	   //	->In logItem Data is stored in []byte.

	   	var content MsgStruct
	   	cont := &MsgStruct{Key:5,Data:"set abc 123"}
	   	t_data, err := json.Marshal(cont)
	   	logitem := &Raft.LogItem{Index: rafttype.GetIndex(), Term: int64(rafttype.Term()), Data:t_data}
	   	err = json.Unmarshal(logitem.Data.([]byte), &content) //decode message into Envelope object.
	           fmt.Println(content)

	   	return
	*/
	DBFILE := "./leveldb2"
	db, err := leveldb.OpenFile(DBFILE+"_"+strconv.Itoa(pid)+".db", nil)
	if err != nil {
		log.Println("err in opening file for leveldb:-", DBFILE, "error is:-", err)
	}
	iter := db.NewIterator(nil, nil)

	var logitem Raft.LogItem
	var content DataType

	for iter.Next() {
		CommitIndex, err := strconv.ParseInt(string(iter.Key()), 10, 64)
		fmt.Println(CommitIndex, err)
		err = json.Unmarshal(iter.Value(), &logitem) //decode message into Envelope object.
		fmt.Println("response of conversion from log to log item", logitem)
		//  now set/get/delete key-value in Map based on Log
		testing := logitem.Data.(map[string]interface{})
		/*
			for key, value := range testing{
				fmt.Println(reflect.TypeOf(value),key,value)
			}
		*/
		fmt.Println(testing["Type"])
		content = DataType{Type: int8(int(testing["Type"].(float64))), Key: testing["Key"].(string), Value: testing["Value"]}
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
	iter.Release()
	log.Println("Error if:-", iter.Error(), "Map is:-", Map)
	defer db.Close()

}

func handle_client(c net.Conn) {
	defer c.Close()

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

		go handleReq(msg)

		//to_send :=string(msg);
		str := strings.Fields(string(msg))

		fmt.Printf("msg=%s\n", msg)
		//fmt.Println(str[0]=="get",str[1])
		action := str[0]
		key := str[1]
		switch {

		case action == "set":
			val := str[2]
			mut.Lock()
			//Map[key] = val
			Set_Val(key, val)
			mut.Unlock()
			fmt.Printf("set Map[%s]=%s", key, val)

		case action == "update":
			val := str[2]
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
			fmt.Println(temp)
			fmt.Printf("get Map[%s]=%s", key, temp)
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

}
