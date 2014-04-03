package KeyValue

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	Raft "github.com/mgkanani/raft"
	"github.com/syndtr/goleveldb/leveldb"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

//	"reflect" //used for identifying the type of interface's data.
)

const (
	SET    = 1
	GET    = 2
	UPDATE = 3
	DELETE = 4

	debug         = false
//	debug         = true
	CONFIG        = "KeyValue.json"
	DBFILE_PREFIX = "./leveldb2"
)

var Map map[string]interface{}

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

var mut sync.RWMutex

var Servers map[int]string

type Message struct {
	Type  int // 0-> set, 3-> get , 1->update,2->delete
	Key   string
	Value string
}

func main() {
	//Read parametes from configuration file.
	file, e := ioutil.ReadFile(CONFIG)
	if e != nil {
		if debug {
			log.Printf("File error: %v\n", e)
		}
		os.Exit(1)
	}
	//decoding file's input and save data into variables.
	var obj jsonobject
	err := json.Unmarshal(file, &obj)
	if err != nil {
		if debug {
			log.Println("error:", err)
		}
	}
	Servers = make(map[int]string)
	for _, value := range obj.Object.Servers {
		Servers[value.MyPid] = value.Socket
	}
	if debug {
		log.Println("Servers for Key-Value:-", Servers)
	}

	if len(os.Args) != 2 {
		fmt.Println("format:-KeyValue pid")
		return
	}
	myid, err := strconv.Atoi(os.Args[1])
	_, present := Servers[myid]
	if err != nil || !present {
		fmt.Println("Please enter valid server id which must be present into", CONFIG, "File. Error is:-", err)
		os.Exit(2)
	}
	socket := Servers[myid]

	Map = make(map[string]interface{}) // create's Map.

	var valid bool
	flagid := &myid

	//Initialize Map.
	ConstructKeyValue(flagid)

	// Initialize raft-server and starts it.
	valid, rafttype = Raft.InitServer(myid, "./config.json", debug) // this will return raft-object and valid= true if it is successful.

	if debug {
		log.Println("Checking whether raft server started status:-", valid)
	}
	if valid {
		http.HandleFunc("/", index)
		err = http.ListenAndServe(socket, nil)
		if err != nil {
			fmt.Println("Address format must be like ipaddr:port_num in config file", CONFIG)
			return
		}
	} else {
		log.Println("error generated in starting server according to configuration file")
		os.Exit(1)
	}
}

func Get_Val(key string) interface{} {
	return Map[key]
	/*	temp,ok:=Map[key]
		if ok{
			return temp
		}else{
			return ""
		}
	*/
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

//currently not used this.
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

//When there is any log-entry from client which should be applied to state-machine or must be logged, this method is called.
//This method must be called only when "rafttype" object is leader because this directly sends log-entry to leader.
func handleReq(msg Raft.DataType) {
	//	if true {
	//Send msg and fetch it's log-entry index.
	index := rafttype.GetIndex(msg)
	if debug {
		log.Println("Log Index for new data is:-", index)
	}
	//Below lines are deprecated.
	/*
		litem := &Raft.LogItem{Index: index, Term: int64(rafttype.Term()), Data: msg}
		if debug {
			log.Println("In handleReq,log-item created is:-", litem)
		}
		//send log-item to state machine.Below line functionality has been deprecated.
		//rafttype.Inbox() <- litem
	*/
	for {
		//wait for response.
		data := <-rafttype.Outbox()
		/*
			if debug{
				log.Println(reflect.TypeOf(data),*(data.(*int64)),index)
			}*/
		//checks whether it the same message which has been replicated successfully.
		if *(data.(*int64)) == index {
			break
		}
	}
	//	}
}

//This should be called at the start of Key-Value server. This is used for saving into Map after reading the log-entries.
func ConstructKeyValue(pid *int) {
	DBFILE := DBFILE_PREFIX + "_" + strconv.Itoa(*pid) + ".db"
	if debug {
		log.Println("Id of server:-", *pid, "\t DB FileName:-", DBFILE)
	}
	db, err := leveldb.OpenFile(DBFILE, nil)
	if err != nil {
		if debug {
			log.Println("err in opening leveldb file:-", DBFILE, "\t error is:-", err)
		}
	}
	//iter := db.NewIterator(nil, nil)

	var logitem Raft.LogItem
	var content Raft.DataType

	temp := make([]byte, 8)
	var i int64
	i = 0
	for {
		i++
		binary.PutVarint(temp, i)
		value, err := db.Get(temp, nil)
		if err != nil {
			if debug{
				log.Println("Error in Get:-", err)
			}
			break
		}
		err = json.Unmarshal(value, &logitem) //decode message into Envelope object.
		if err != nil {
			if debug {
				log.Println("In InitServer of Raft,Error during Marshaling:-", err)
			}
		}
		if debug {
			log.Println("converted to log-item from bytes:-", logitem)
		}
		if logitem.Data != nil {
			testing := logitem.Data.(map[string]interface{})
			/*
				// uncomment to see the type of each field of testing variable.
				for key, value := range testing{
					fmt.Println(reflect.TypeOf(value),key,value)
				}
				//log.Println(testing["Type"])
			*/
			content = Raft.DataType{Type: int8(int(testing["Type"].(float64))), Key: testing["Key"].(string), Value: testing["Value"]}
			if debug {
				//log.Println("created object from log-entry is:-", content)
			}
			switch content.Type {
			case SET, UPDATE: //set value
				if content.Value != nil {
					Map[content.Key] = content.Value.(string)
				}
				break
			case DELETE: //delete value
				delete(Map, content.Key)
				break
			}
		}
	}
	/*	iter.Release()
		if iter.Error() != nil {
			if debug {
				log.Println("Error during iteration on data of log-entries:-", iter.Error())
			}
		}
	*/
	if debug {
		PrintMap()
	}
	defer db.Close()
}

func PrintMap() {
	if debug {
		log.Println("\n\nThe content of Map-variable is:-\n", Map, "\n\n\n\n")
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var is_leader int
		rafttype.Leader(&is_leader)
		//is_leader will have >0 if it is leader other wise it will send 0 for follower,-1 for candidate.

		if is_leader > 0 {
			var msgs Raft.DataType
			body, e := ioutil.ReadAll(r.Body)
			if e != nil {
				if debug {
					log.Println("error @KeyValue.go: in 'index' func :if : ReadAll", e)
				}
			}
			e = json.Unmarshal(body, &msgs)

			if e != nil {
				if debug {
					log.Println("error @KeyValue.go: in 'index' func :if :Unmarshal", e)
				}
			}
			if debug {
				log.Println("client request has been received on:-", r.Host, "\t message is :-", msgs)
			}

			switch msgs.Type {
			case SET:
				if debug {
					fmt.Println(" In SET case:-")
				}
				handleReq(msgs)
				mut.Lock()
				//Map[key] = val
				Set_Val(msgs.Key, msgs.Value)
				mut.Unlock()
				if debug {
					log.Printf("\nSet:: Map[%s]=%s\n", msgs.Key, msgs.Value)
				}

			case UPDATE: //Update
				if debug {
					fmt.Println(" In UPDATE case:-")
				}
				handleReq(msgs)
				mut.Lock()
				//Map[key] = val
				Update_Val(msgs.Key, msgs.Value)
				mut.Unlock()
				if debug {
					log.Printf("\nUpdate:: Map[%s]=%s\n", msgs.Key, msgs.Value)
				}

			case GET: //Get
				if debug {
					fmt.Println(" In GET case:-")
				}
				mut.RLock()
				temp := Get_Val(msgs.Key)
				mut.RUnlock()
				if temp != nil {
					data := temp.(string)
					fmt.Fprintf(w, data)
					if debug {
						log.Printf("Sent data to client in reply is:-", temp)
					}
				} else {
					fmt.Fprintf(w, "")
				}

			case DELETE: //Delete
				if debug {
					fmt.Println(" In DELETE case:-")
				}
				//msg := Raft.DataType{Type: DELETE, Key: key}
				handleReq(msgs)
				mut.Lock()
				Delete(msgs.Key)
				mut.Unlock()
				if debug {
					log.Println("\nDeleted key from map is ", msgs.Key)
				}
			}
		} else {
			res := rafttype.Vote()
			new_url := "http://" + Servers[res] + "/"
			var msgs Raft.DataType
			body, e := ioutil.ReadAll(r.Body)
			json.Unmarshal(body, &msgs)
			if debug {
				fmt.Println("Request recieved at:-", r.Host, ",redirected to:-", new_url, "\n Server has voted to :-", res, "Data is:-", msgs, "Error status:-", e)
			}
			http.Redirect(w, r, new_url, 303)
		}
	}
}
