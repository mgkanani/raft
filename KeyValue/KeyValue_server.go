package main

import (
	"encoding/json"
	//	"flag"
	"fmt"
	Raft "github.com/mgkanani/raft"
	"github.com/syndtr/goleveldb/leveldb"
	//	"io"
	"io/ioutil"
	"log"
	//	"net"
	"os"
	"strconv"
	//	"strings"
	"net/http"
	"sync"

//	"reflect"
)

const (
	SET    = 0
	GET    = 3
	UPDATE = 1
	DELETE = 2
)

var Map map[string]interface{}
var Pending map[int64]bool

type MsgStruct struct {
	Key  int64
	Data interface{}
}

type jsonobject struct {
	Object ServersType
}

type ServersType struct {
	Servers []ServerType
}
type ServerType struct {
	Socket string //used to storing socket string for specified pid.
	MyPid  int    //stores Pid.
}

var rafttype *Raft.RaftType

var myid = 0

//var mut sync.Mutex
var mut sync.RWMutex

var Servers map[int]string

type Message struct {
	Type  int // 0-> set, 3-> get , 1->update,2->delete
	Key   string
	Value string
}

/*
func index(w http.ResponseWriter, r *http.Request) {
        http.Redirect(w , r , "http://10.5.0.26:80/", http.StatusMovedPermanently)

        var person Message
        body, e := ioutil.ReadAll(r.Body)
        json.Unmarshal(body, &person)

        fmt.Println(person,e)
        response, _ := json.Marshal(person)
        fmt.Fprintf(w, string(response))

}
*/

func main() {

	file, e := ioutil.ReadFile("KeyValue.json")
	if e != nil {
		fmt.Printf("File error: %v\n", e)
		os.Exit(1)
	}
	//decoding file's input and save data into variables.
	var obj jsonobject
	err := json.Unmarshal(file, &obj)
	if err != nil {
		fmt.Println("error:", err)
	}
	Servers = make(map[int]string)
	for _, value := range obj.Object.Servers {
		Servers[value.MyPid] = value.Socket
	}
	fmt.Println(Servers)

	if len(os.Args) != 2 {
		fmt.Println("format:-KeyValue pid")
		return
	}
	myid, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	socket := Servers[myid]

	Map = make(map[string]interface{})
	Pending = make(map[int64]bool)

	/*
		addr_port, err := net.ResolveTCPAddr("tcp", socket)
		ln, err := net.ListenTCP("tcp", addr_port)
		if err != nil {
			fmt.Println("use different port number, given port already in use.\nin", err.Error())
			return
		}

		fmt.Println("Listening at addr:port", addr_port)
		defer ln.Close()
	*/
	var valid bool
	flagid := &myid

	ConstructKeyValue(flagid)

	valid, rafttype = Raft.InitServer(myid, "./config.json", true)

	fmt.Println("valid:-", valid)
	if valid {
		/*		var is_leader int
				for {
					conn, err := ln.Accept()
					rafttype.Leader(&is_leader)
					fmt.Println("is_leader:-", is_leader)
					if myid == is_leader {
						go handle_client(conn)
					} else {
						_, _ = conn.Write([]byte("Please contact Server -" + Servers[rafttype.Vote()]))
						conn.Close()
					}
				}
		*/

		http.HandleFunc("/", index)
		err = http.ListenAndServe(socket, nil)
		if err != nil {
			fmt.Println("Address format is ipaddr:port_num")
			return
		}

	} else {
		log.Println("error generated in starting server according to configuration file")
		os.Exit(1)
	}
}

func Get_Val(key string) interface{} {
	return Map[key]
}

func Set_Val(key string, val interface{}) {
	Map[key] = val
	return
}

func Update_Val(key string, val interface{}) {
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
				if content.Value != nil {
					Map[content.Key] = content.Value.(string)
				}
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

//func handle_client(c net.Conn) {
func index(w http.ResponseWriter, r *http.Request) {
	var is_leader int
	rafttype.Leader(&is_leader)

	//fmt.Println("Leader is:-",is_leader,"myid:-",myid)
	if is_leader > 0 {

		//	for {

		//var msgs Message
		var msgs Raft.DataType
		body, e := ioutil.ReadAll(r.Body)
		json.Unmarshal(body, &msgs)

		fmt.Println(msgs, e)
		fmt.Println("client request has been received on:-", r.Host, msgs, e)
		//        response, _ := json.Marshal(person)
		//        fmt.Fprintf(w, string(response))

		/*		msg := make([]byte, 1000)

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
		*/
		switch msgs.Type {

		case 0: //SET
			//val := str[2]
			//msg := msgs//Raft.DataType{Type: SET, Key: key, Value: val}
			handleReq(msgs)
			mut.Lock()
			//Map[key] = val
			Set_Val(msgs.Key, msgs.Value)
			mut.Unlock()
			fmt.Printf("set Map[%s]=%s", msgs.Key, msgs.Value)

		case 1: //Update
			//val := str[2]
			//msg := Raft.DataType{Type: UPDATE, Key: key, Value: val}
			handleReq(msgs)
			mut.Lock()
			//Map[key] = val
			Update_Val(msgs.Key, msgs.Value)
			mut.Unlock()
			//fmt.Printf("update Map[%s]=%s", str[1], str[2])

		case 3: //Get
			//fmt.Printf("s Map[%s]=%s e\n", str[1], Map[key])
			//fmt.Printf("s Map[%s]=%s e\n", str[1], Map["abc"])
			//temp:=string(strings.TrimSpace(str[1]));
			mut.RLock()
			temp := Map[msgs.Key].(string)
			//temp := []byte(Get_Val(key))
			//temp := []byte(Get_Val(key))
			mut.RUnlock()
			//fmt.Println(temp)
			//fmt.Printf("get Map[%s]=%s", key, temp)
			//n, err = c.Write(temp)
			fmt.Fprintf(w, temp)
			/*			if n == 0 {
							n, err = c.Write([]byte("\n"))
						}

						if err != nil {
							fmt.Printf("ERROR: write\n")
							fmt.Print(err)
							return
						}
						fmt.Printf("SERVER: sent %v bytes\n", n)
			*/

		case 2: //Delete
			//msg := Raft.DataType{Type: DELETE, Key: key}
			handleReq(msgs)
			mut.Lock()
			Delete(msgs.Key)
			mut.Unlock()
			fmt.Printf("delete Map[%s]", msgs.Key)
			/*
				case action == "rename":
					keyfrom := str[1]
					keyto := str[2]
					mut.Lock()
					Rename(keyfrom, keyto)
					mut.Unlock()
					fmt.Printf("Rename from Map[%s] to Map[%s]", keyfrom, keyto)
			*/
		}
		//	}
	} else {
		//new_url:="http://"+Servers[res]+"/"
		//fmt.Println("client request has been received on:-",r.Host,",redirected to:-",new_url,"\n Server has voted to :-",res);

		res := rafttype.Vote()
		new_url := "http://" + Servers[res] + "/"
		var msgs Raft.DataType
		body, e := ioutil.ReadAll(r.Body)
		json.Unmarshal(body, &msgs)
		fmt.Println("client request has been received on:-", r.Host, ",redirected to:-", new_url, "\n Server has voted to :-", res, "Data is:-", msgs, e)

		http.Redirect(w, r, new_url, 303)
		//fmt.Println("error:-",e)
	}
}
