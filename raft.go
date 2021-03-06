package raft

import (
	"encoding/binary"
	"encoding/json"
	//"fmt"
	cluster "github.com/mgkanani/cluster"
	"github.com/syndtr/goleveldb/leveldb"
	"log"
	rand "math/rand"
	"strconv"
	"sync"
	"time"
)

const (
	DBFILE = "./leveldb2"

	FOLLOWER  = 0
	CANDIDATE = 1
	LEADER    = 2

	REQ     = 0
	REP     = 1
	APP     = 2
	L_I_REQ = 3 //LogItem request received at leader. only used for leader.
	HEART   = 4 //used to identify heartbeat message from Leader.
	AER     = 5 //Append Entry reply type
	L_I_REP = 6 //LogItem response received.
)

var debug = true
var db *leveldb.DB

type RaftType struct {
	serv *Server
}

//used for Key value purpose only. This structure is used in KeyValue_server.go file to create objects.
type DataType struct {
	Type  int8 //0 to set, 1 to update,2 to delete.
	Key   string
	Value interface{}
}

type Raft interface {
	Term() int   //returns the current term number.
	Leader() int //returns id of a leader if there exist, otherwise returns -1;

	// Mailbox for state machine layer above to send commands of any
	// kind, and to have them replicated by raft.  If the server is not
	// the leader, the message will be silently dropped.
	Outbox() chan<- interface{}

	//Mailbox for state machine layer above to receive commands. These
	//are guaranteed to have been replicated on a majority
	Inbox() <-chan *LogItem
	//Inbox() <-chan interface{}

	//Remove items from 0 .. index (inclusive), and reclaim disk
	//space. This is a hint, and there's no guarantee of immediacy since
	//there may be some servers that are lagging behind).

	DiscardUpto(index int64)
}

//This function has not been used.
func (ser *Server) DiscardUpto(index int64) {
	//discards all log entries after given index means >Index will be droped for follower only.
	if ser.ServState.my_state == FOLLOWER {
		for k := int64(len(ser.ServState.Log)); k > index; k-- {
			log.Println("deleting entry", k, ser.ServState.Log[k])
			delete(ser.ServState.Log, k)
		}

	}

}

//Returns the channel for Outbox
func (raft *RaftType) Outbox() chan interface{} {
	return raft.serv.out
}

//Returns the channel for Inbox
func (raft *RaftType) Inbox() chan *LogItem {
	return raft.serv.in
}

//Prints the whole data except Log of Server Object.
func (ser *Server) PrintData() {
	log.Println("ServerID", ser.ServerInfo.MyPid, " \t Term:-", ser.ServState.my_term, "\t State:-", ser.ServState.my_state, "\nCommitIndex:-", ser.ServState.CommitIndex, "\t LastApplied:-", ser.ServState.LastApplied, "\nFollowers:-", ser.ServState.followers, "\t NextIndex:-", ser.ServState.NextIndex, "\t MatchIndex:-", ser.ServState.MatchIndex)
}

//Prints the whole data of Server Object.
func (ser *Server) PrintDataLog() {
	log.Println("ServerID", ser.ServerInfo.MyPid, " \t Term:-", ser.ServState.my_term, "\t State:-", ser.ServState.my_state, "\t Log:-", ser.ServState.Log, "\nCommitIndex:-", ser.ServState.CommitIndex, "\t LastApplied:-", ser.ServState.LastApplied, "\nFollowers:-", ser.ServState.followers, "\t NextIndex:-", ser.ServState.NextIndex, "\t MatchIndex:-", ser.ServState.MatchIndex)
}

//This will retrieve msg and create Logitem object then apply it to Log data structure. This is used to ensure that each new log entry has different logindex.
func (rt *RaftType) GetIndex(msg DataType) int64 {
	//to ensure that each request have different Index.
	if debug {
		log.Println("In GetIndex method, Message is:-", msg)
	}
	rt.serv.ServState.Log[rt.serv.ServState.LastApplied+1] = LogItem{Index: rt.serv.ServState.LastApplied + 1, Term: int64(rt.serv.ServState.my_term), Data: msg}
	rt.serv.ServState.LastApplied += 1

	return rt.serv.ServState.LastApplied
}

//this function has been deprecated. Currently not in use.
func (raft *RaftType) handleInbox() {
	if debug {
		//log.Println("In inbox")
	}
	for {
		//wait for log entry request. This step can be eliminited because we have already fetched it.
		req := <-raft.serv.in //Log Item received. and req.Data must be in []byte
		if debug {
			log.Println("New Entry Received in handleInbox:-", req)
		}
		t_data, err := json.Marshal(req)
		if err != nil {
			if debug {
				log.Println("In handleInbox:- Marshaling error: ", err)
			}
		} else {
			var envelope cluster.Envelope
			data := string(t_data)
			// braodcast the requestFor vote.
			if raft.serv.ServState.my_state == LEADER { //If it is leader.
				envelope = cluster.Envelope{Pid: raft.serv.ServerInfo.Pid(), MsgId: L_I_REQ, Msg: data}
			} else if raft.serv.ServState.my_state == CANDIDATE {
			} else {
				envelope = cluster.Envelope{Pid: raft.serv.ServState.vote_for, MsgId: L_I_REQ, Msg: data}
			}
			raft.serv.ServerInfo.Inbox() <- &envelope
		}
	}
}

// Identifies an entry in the log
type LogItem struct {
	// An index into an abstract 2^64 size array
	Index int64

	Term int64

	// The data that was supplied to raft's inbox
	Data interface{}
}

//Msg Type,whether it is request or reply.
type MsgType struct {
	MType byte        // 1 for reply, 0 for request.
	Msg   interface{} //actual object/message.
}

type AppendEntries struct {
	Term         int
	LeaderId     int
	PrevLogIndex int64
	PrevLogTerm  int64
	//	Entries          map[int64]LogItem// this will not work with json because of int64 as a key in map.
	Entries           LogItem
	LeaderCommitIndex int64
}

type AE_Reply struct {
	//Reply strucrure.
	Term          int
	Success       bool
	PrevLogIndex  int64
	ExpectedIndex int64
}

type HeartBeat struct {
	Term             int
	LeaderId         int
	PrevLogIndex     int64
	PrevLogTerm      int64
	LeaderCommitIdex int64
}

type Reply struct {
	//Reply strucrure.
	Term   int
	Result bool // true if follower has voted, otherwise false for rejection for vote.
}

type Request struct {
	//Request message structure.
	Term         int //Request for vote for this Term.
	CandidateId  int //Requested candidate-id.
	LastLogIndex int64
	LastLogTerm  int64
}

// Server State data structure
type ServerState struct {
	my_term   int //default will be zero
	vote_for  int //value will be pid of leader.
	my_state  int
	followers map[int]int
	Log       map[int64]LogItem

	CommitIndex int64 //index of highest logentry known to be committed.
	LastApplied int64 // index of highest log entry applied to state machine.

	//for each server index of the
	NextIndex  map[int]int64 // next log entry to send
	MatchIndex map[int]int64 // highest log entry known to be replicated on server
}

type Server struct {
	ServState  ServerState        //Server-State informtion are stored.
	ServerInfo cluster.ServerType //Server meta information will be stored like ip,port.
	in         chan *LogItem      //for input channel(Inbox)
	out        chan interface{}   //for output channel(OutBox)

}

//Update the Term
func (serv *ServerState) UpdateTerm(term int) {
	if debug {
		log.Println("Term :-", term)
	}
	serv.my_term = term
}

//Server update it's variable for which it has voted.
func (serv *ServerState) UpdateVote_For(pid int) {
	serv.vote_for = pid
	if debug {
		//log.Println("Voted for :-", pid , "for term:-",serv.my_term)
	}
}

//Server updates it's state and returns true if it is successful, otherwise false.
func (serv *ServerState) UpdateState(new_state int) bool {
	if new_state > 2 || new_state < 0 {
		return false
	} else {
		serv.my_state = new_state
		if debug {
			//	log.Println("New State :-", new_state)
		}
		//fmt.Println("stateUpdated to:-",new_state,"for",serv.ServerInfo.MyPid,serv)
		return true
	}
}

//returns the current Term number.
func (serv Server) Cur_Term() int {
	return serv.ServState.my_term
}

//returns the Server-id for which server it has voted.
func (serv Server) Vote() int {
	return serv.ServState.vote_for
}

//returns the Server-id for which server it has voted.
func (serv RaftType) Vote() int {
	return serv.serv.ServState.vote_for
}

//returns the current term, To be used from outside of this package.
func (rt *RaftType) Term() int {
	if debug {
		//log.Println("Asked for Term :-", rt.serv.ServState.my_term,rt.serv.ServerInfo.Pid())
	}
	return rt.serv.Cur_Term()
}

//to check whether itself is leader, To be used from outside of this package.
func (rt *RaftType) Leader(id *int) {
	//fmt.Println("In Leader function:-",rt,*id)
	if rt.serv.ServState.my_state == FOLLOWER {
		//return rt.serv.ServState.vote_for
		*id = 0
	} else if rt.serv.ServState.my_state == LEADER {
		*id = rt.serv.ServState.vote_for
	} else {
		*id = -1
	}
	//fmt.Println("In Leader function:-",rt,*id)
}

//Used in InitServer method.
func (rt *RaftType) setServer(serv *Server) {
	rt.serv = serv
}

//Initializes the servers with given parameters.
func InitServer(pid int, file string, dbg bool) (bool, *RaftType) {
	/*	fle, err := os.OpenFile("log_pid_"+strconv.Itoa(pid) ,os.O_WRONLY | os.O_CREATE | os.O_APPEND, 0666)
		if err != nil {
		    println("error opening file: %v", err)
		}
		defer fle.Close()

		log.SetOutput(fle)
	*/
	debug = dbg
	serv := new(Server)
	rtype := RaftType{}
	rtype.setServer(serv)

	serv.in, serv.out = make(chan *LogItem), make(chan interface{})
	serv.ServerInfo = cluster.New(pid, file)
	serv.ServState.followers = make(map[int]int)
	serv.ServState.Log = make(map[int64]LogItem)
	serv.ServState.NextIndex = make(map[int]int64)
	serv.ServState.MatchIndex = make(map[int]int64)

	if serv.ServerInfo.Valid {
		var err error
		db, err = leveldb.OpenFile(DBFILE+"_"+strconv.Itoa(pid)+".db", nil)
		if err != nil {
			if debug {
				log.Println("err in opening file for leveldb:-", DBFILE, "error is:-", err)
			}
		} else {
			// Here iterator is not used to fetch log-entries, Iterator does not gives the log-entries in the same order in which we have inserted. We know that we have inserted in increasing order of Log-Index so fetch in the same order. It will come out as soon as it finds any missing.
			var logitem LogItem
			temp := make([]byte, 8)
			var i int64
			i = 0
			for {
				i++
				binary.PutVarint(temp, i) //int64 to []byte conversion
				value, err := db.Get(temp, nil)
				if err != nil {
					if debug {
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
				serv.ServState.Log[i] = logitem //Assign log-item to Log
			}
			// Initialize Commitindex and Last applied index
			serv.ServState.CommitIndex = i - 1
			serv.ServState.LastApplied = i - 1
			if debug {
				log.Println("LastApplied:-", serv.ServState.LastApplied, "Comiit:-", serv.ServState.CommitIndex, "serv.ServState.Log[int64(len(serv.ServState.Log))].Index:-", serv.ServState.Log[int64(len(serv.ServState.Log))].Index, "serv.ServState.Log[serv.ServState.LastApplied].Term", serv.ServState.Log[serv.ServState.LastApplied].Term)
			}
			/*//Before above this was used.
			iter := db.NewIterator(nil, nil)
			for iter.Next() {
				// Remember that the contents of the returned slice should not be modified, and
				// only valid until the next call to Next.
				//converts []byte to int64
				serv.ServState.CommitIndex, _ = binary.Varint(iter.Key())
				err = json.Unmarshal(iter.Value(), &logitem) //decode message into Envelope object.
				if err!=nil{
					if debug {
						log.Println("In InitServer of Raft,Error during Marshaling:-",err)
					}
				}
				serv.ServState.Log[serv.ServState.CommitIndex] = logitem
			}
			serv.ServState.LastApplied = serv.ServState.CommitIndex
			iter.Release()
			//log.Println("Error if:-", iter.Error(), serv.ServState.LastApplied)
			*/
		}
		if debug {
			serv.PrintData()
		}
		go serv.start()
		//go rtype.handleInbox() //not in use currently
		//go rtype.handleOutbox() //not in use currently
	}
	return serv.ServerInfo.Valid, &rtype //return true if server initialization succeed otherwise false, and also return the raft object.
}

//starts the leader election process.
func (serv *Server) start() {
	var mutex = &sync.Mutex{}
	for {
		if debug {
			//log.Println("Inside of for loop of start method ")
		}

		switch serv.ServState.my_state {
		case FOLLOWER:
			//follower
			//fmt.Println("Follower : Serverid-", serv.ServerInfo.MyPid)
			serv.StateFollower(mutex)
		case CANDIDATE:
			//candidate
			//fmt.Println("Candidate : Serverid-", serv.ServerInfo.MyPid)
			serv.StateCandidate(mutex)
		case LEADER:
			//leader
			//fmt.Println("Leader : Serverid-", serv.ServerInfo.MyPid)
			serv.StateLeader(mutex)
		}
	}
	if debug {
		log.Println("=========================Outside of for loop")
	}

	defer db.Close()
}

//Server goes to Follower state.
func (serv *Server) StateFollower(mutex *sync.Mutex) {
	//duration := 600 * time.Millisecond
	//duration := time.Duration((rand.Intn(50)+serv.ServerInfo.MyPid*60)*12) * time.Millisecond
	duration := 500*time.Millisecond + time.Duration(rand.Intn(600))*time.Millisecond
	timer := time.NewTimer(duration)

	select { //used for selecting channel for given event.
	case enve := <-serv.ServerInfo.Inbox():
		//new message has arrived.

		switch enve.MsgId {
		case REQ:
			//Request Rcvd.
			var req Request
			err := json.Unmarshal([]byte(enve.Msg.(string)), &req) //decode message into Envelope object.
			if err != nil {                                        //error into parsing/decoding
				if debug {
					log.Println("Follower: Unmarshaling error:-\t", err)
				}
			}
			//fmt.Println("Follower : Serverid-", serv.ServerInfo.MyPid, " Request is:-.", req)

			if debug { //to check during debuging which branch has been selected.
				log.Println("In follower:-", req.LastLogIndex, ">=", serv.ServState.LastApplied, " &&", req.Term, " > ", serv.Cur_Term(), "&&", req.LastLogTerm, ">=", serv.ServState.Log[serv.ServState.LastApplied].Term)
				log.Println("In follower:- !(", enve.Pid, "==", serv.Vote(), "&&", req.Term, "==", serv.Cur_Term(), ")")
			}

			if req.Term < serv.Cur_Term() { //neglect message.
				/*
				 */
			} else if req.LastLogIndex >= serv.ServState.LastApplied && req.Term > serv.Cur_Term() && req.LastLogTerm >= serv.ServState.Log[serv.ServState.LastApplied].Term {
				////////////  request-accept branch //////////
				//higher or equal term received, reset timer,send accept for request.
				timer.Reset(duration) //reset timer
				var reply *Reply
				reply = &Reply{Term: req.Term, Result: true}
				t_data, err := json.Marshal(reply)
				if err != nil {
					if debug {
						log.Println("Follower:- request-accept branch:- Marshaling error: ", err)
					}
				}
				data := string(t_data)
				envelope := cluster.Envelope{Pid: enve.Pid, MsgId: 1, Msg: data}
				mutex.Lock()
				serv.ServState.UpdateVote_For(req.CandidateId)
				serv.ServState.UpdateTerm(req.Term)
				if debug {
					//log.Println("Updated Term for :-", serv.ServerInfo.MyPid)
				}
				mutex.Unlock()
				serv.ServerInfo.Outbox() <- &envelope
				if debug {
					log.Println("Follower :- Accept-Reply sent :-", reply)
				}

			} else if !(enve.Pid == serv.Vote() && req.Term == serv.Cur_Term()) { //getting request for vote,reject the request
				////////////  request-reject branch //////////
				reply := &Reply{Term: req.Term, Result: false}
				data, err := json.Marshal(reply)
				if err != nil {
					if debug {
						log.Println("Follower:- request-reject branch :- Marshaling error: ", err)
					}
				}
				serv.ServState.UpdateVote_For(req.CandidateId)
				envelope := cluster.Envelope{Pid: enve.Pid, MsgId: 1, Msg: string(data)}
				serv.ServerInfo.Outbox() <- &envelope
				if debug {
					log.Println("Follower :- Reject-Reply sent :-", reply)
				}
			} /*else if req.Term > serv.Cur_Term(){//wrong candidate has sent request.
			                	        timer.Stop() //stop timer.
				                        mutex.Lock()
			        	                serv.ServState.UpdateVote_For(serv.ServerInfo.MyPid) //giving him self vote.
			                	        _ = serv.ServState.UpdateState(1)                    // update state to be a candidate.
				                        serv.ServState.UpdateTerm(req.Term + 1)       //increment term by one.
			        	                mutex.Unlock()
			                	        var req *Request
			                        	req = &Request{Term: serv.Cur_Term(), CandidateId: serv.ServerInfo.Pid(),LastLogIndex:serv.ServState.Log[int64(len(serv.ServState.Log))].Index,LastLogTerm:serv.ServState.Log[int64(len(serv.ServState.Log))].Term}
				                        t_data, err := json.Marshal(req)
			        	                if err != nil {
			                	                if debug {
			                        	                log.Println("Follower:- After Awaking :- Marshaling error: ", err)
			                                	}
				                        } else {
			        	                        data := string(t_data)
			                	                // braodcast the requestFor vote.
			                        	        envelope := cluster.Envelope{Pid: cluster.BROADCAST, MsgId: 0, Msg: data}
			                                	serv.ServerInfo.Outbox() <- &envelope
				                                //fmt.Println("Follower : Serverid-", serv.ServerInfo.MyPid, "After Awaking,req sent for vote", req, "actual sent", data, envelope, serv)
			        	                }
				                        return
			}*/

			break

		case HEART:
			//HeartBeat rcvd.
			timer.Reset(duration) //reset timer
			var hrt HeartBeat
			err := json.Unmarshal([]byte(enve.Msg.(string)), &hrt) //decode message into Envelope object.
			if err != nil {                                        //error into parsing/decoding
				if debug {
					log.Println("Follower,HeartBeat-Received: Unmarshaling error:-\t", err)
				}
			}
			if debug {
				log.Println("Follower:- Heartbeat Recvd for ", "sid-", serv.ServerInfo.MyPid, "from", enve.Pid, hrt)
			}
			// send positive reply.
			var reply *Reply
			reply = &Reply{Term: hrt.Term, Result: true}
			t_data, err := json.Marshal(reply)
			if err != nil {
				if debug {
					log.Println("Follower,HeartBeat:- Marshaling error: ", err)
				}
			}
			data := string(t_data)
			envelope := cluster.Envelope{Pid: enve.Pid, MsgId: 1, Msg: data}
			serv.ServerInfo.Outbox() <- &envelope
			if debug {
				log.Println("Follower :- Heartbeat Reply sent :-", reply)
			}
			break

		case APP:
			// request for Append entry recieved.
			timer.Reset(duration)

			var app AppendEntries
			err := json.Unmarshal([]byte(enve.Msg.(string)), &app) //decode message into Envelope object.
			if err != nil {                                        //error into parsing/decoding
				if debug {
					log.Println("Follower,AppendEntries: Unmarshaling error:-\t", err)
				}
			}
			if debug {
				//log.Println("Append entry Recvd for ", "sid-", serv.ServerInfo.MyPid, "from", enve.Pid, app, serv.ServState.Log, serv.ServState.LastApplied, serv.ServState.CommitIndex)
				log.Println("Append Entry Recvd:-", app)
			}
			// send positive reply.
			var reply *AE_Reply
			//new  logic starts
			if app.PrevLogIndex == 0 && app.PrevLogTerm == 0 { // to handle very basic case when servers are just started and have no data at all.
				if debug {
					log.Println("===========First entry came:-", app)
				}
				serv.ServState.Log[1] = app.Entries
				serv.ServState.LastApplied = 1
				//serv.ServState.CommitIndex = app.LeaderCommitIndex
				t_data := make([]byte, 8)
				binary.PutVarint(t_data, serv.ServState.LastApplied) //int64 to []byte conversion.
				t_data1, err := json.Marshal(app.Entries)
				if err != nil && debug {
					log.Println("In Follwer:APP:-", err)
				}
				err = db.Put(t_data, t_data1, nil) //writting to database.
				if err != nil {
					if debug {
						log.Println("In Follwer:APP writing to db error.:-", err)
					}
				}
				reply = &AE_Reply{Term: app.Term, Success: true, PrevLogIndex: app.PrevLogIndex, ExpectedIndex: serv.ServState.LastApplied + 1}

			} else if app.Term < serv.ServState.my_term {
				//discard msg since it is of no use.
			} else if serv.ServState.Log[serv.ServState.LastApplied].Term != app.PrevLogTerm || serv.ServState.LastApplied != app.PrevLogIndex {

				/////////////// Unmatched-Append /////////////////////

				if debug {
					log.Println("Follower:- In Unmatched-Append entry brach ", serv.ServState.Log[serv.ServState.LastApplied].Term, "!=", app.PrevLogTerm, "OR", serv.ServState.LastApplied, "!=", app.PrevLogIndex)
				}

				if serv.ServState.LastApplied != app.PrevLogIndex {

					if debug {
						log.Println("Unmatched Append entry Recvd for ", serv.ServState.Log[serv.ServState.LastApplied].Term, "!=", app.PrevLogTerm, "OR", serv.ServState.LastApplied, "!=", app.PrevLogIndex)
					}

					reply = &AE_Reply{Term: app.Term, Success: false, PrevLogIndex: app.PrevLogIndex, ExpectedIndex: serv.ServState.LastApplied + 1}

					if debug {
						log.Println("Unmatched Append entry Recvd for ", "received.PrevLogIndex:-", app.PrevLogIndex, "LastApplied:-", serv.ServState.LastApplied, serv.ServState.CommitIndex, "Expected:-", reply.ExpectedIndex)
					}

				} else { //less than

					if debug {
						log.Println("Unmatched Append entry Recvd for ", serv.ServState.Log[serv.ServState.LastApplied].Term, "!=", app.PrevLogTerm, "OR", serv.ServState.LastApplied, "!=", app.PrevLogIndex)
					}

					_, exist := serv.ServState.Log[serv.ServState.LastApplied-1]
					if exist {

						if debug {
							log.Println("Server Data", serv.ServState.Log, serv.ServState.LastApplied, serv.ServState.CommitIndex)
						}

						serv.ServState.LastApplied -= 1
					}
					reply = &AE_Reply{Term: app.Term, Success: false, PrevLogIndex: app.PrevLogIndex, ExpectedIndex: serv.ServState.LastApplied + 1}

					if debug {
						log.Println("Unmatched Append entry Recvd for ", "received.PrevLogIndex:-", app.PrevLogIndex, "LastApplied:-", serv.ServState.LastApplied, serv.ServState.CommitIndex, "Expected:-", reply.ExpectedIndex)
					}

				}
			} else {
				/////////////// Matched-Append /////////////////////
				if debug {
					log.Println("Follower:- In Matched-Append entry brach ", serv.ServState.Log[serv.ServState.LastApplied].Term, "!=", app.PrevLogTerm, "OR", serv.ServState.LastApplied, "!=", app.PrevLogIndex)
				}

				serv.ServState.Log[app.PrevLogIndex+1] = app.Entries
				serv.ServState.LastApplied = app.PrevLogIndex + 1
				serv.ServState.CommitIndex = app.LeaderCommitIndex
				//t_data, err := json.Marshal(serv.ServState.LastApplied)
				t_data := make([]byte, 8)
				binary.PutVarint(t_data, serv.ServState.LastApplied)
				t_data1, err := json.Marshal(app.Entries)
				if err != nil && debug {
					log.Println("In Follwer:APP:-", err)
				}
				//err = db.Delete(t_data, nil)       //writting to database.
				err = db.Put(t_data, t_data1, nil) //writting to database.
				if err != nil && debug {
					log.Println("In Follwer:APP writing to db error.:-", err)
				}
				reply = &AE_Reply{Term: app.Term, Success: true, PrevLogIndex: app.PrevLogIndex, ExpectedIndex: serv.ServState.LastApplied + 1}
				if debug {
					log.Println("matched Append entry Recvd for ", "received.PrevLogIndex:-", app.PrevLogIndex, "LastApplied:-", serv.ServState.LastApplied, serv.ServState.CommitIndex, "Expected:-", reply.ExpectedIndex)
				}
			}
			// new logic ends
			t_data, err := json.Marshal(reply)
			if err != nil {
				if debug {
					log.Println("Follower,AppendEntries:- Marshaling error: ", err)
				}
			}
			data := string(t_data)
			envelope := cluster.Envelope{Pid: enve.Pid, MsgId: AER, Msg: data}
			serv.ServerInfo.Outbox() <- &envelope
			if debug {
				log.Println("In Follwer:APP Request's reply sent:-", reply)
			}

			time.Sleep(100 * time.Millisecond)
			break

		}

	case <-timer.C:

		if debug {
			log.Println("Timeout for follower:-", serv.ServerInfo.MyPid, "Duration was:-", duration)
		}
		//declare leader has gone.
		serv.ServState.UpdateVote_For(0) //leader may crashed.
		//wait for 150-300ms
		sleep_time := time.Duration(150 + rand.Intn(151))
		sleep_time *= time.Millisecond
		time.After(sleep_time) //sleeps for random time in between 150ms and 300 ms.
		//fmt.Println(" awaken")
		if serv.Vote() == 0 { //still no candidate exist.
			mutex.Lock()
			serv.ServState.UpdateVote_For(serv.ServerInfo.MyPid) //giving him self vote.
			_ = serv.ServState.UpdateState(1)                    // update state to be a candidate.
			serv.ServState.UpdateTerm(serv.Cur_Term() + 1)       //increment term by one.
			mutex.Unlock()
			var req *Request
			req = &Request{Term: serv.Cur_Term(), CandidateId: serv.ServerInfo.Pid(), LastLogIndex: serv.ServState.Log[int64(len(serv.ServState.Log))].Index, LastLogTerm: serv.ServState.Log[int64(len(serv.ServState.Log))].Term}
			t_data, err := json.Marshal(req)
			if err != nil {
				if debug {
					log.Println("Follower:- After Awaking :- Marshaling error: ", err)
				}
			} else {
				data := string(t_data)
				// braodcast the requestFor vote.
				envelope := cluster.Envelope{Pid: cluster.BROADCAST, MsgId: 0, Msg: data}
				serv.ServerInfo.Outbox() <- &envelope
				//fmt.Println("Follower : Serverid-", serv.ServerInfo.MyPid, "After Awaking,req sent for vote", req, "actual sent", data, envelope, serv)
			}
			timer.Stop() //stop timer.
			return
		}

	}
	timer.Stop()

}

//Server goes to Candidate state.
func (serv *Server) StateCandidate(mutex *sync.Mutex) {

	//duration := 1*time.Second + time.Duration(rand.Intn(151))*time.Millisecond
	//duration := 150*time.Millisecond + time.Duration(rand.Intn(151))*time.Millisecond //choose duration between 150-300ms.
	duration := 450*time.Millisecond + time.Duration(rand.Intn(151))*time.Millisecond //choose duration between 150-300ms.
	timer := time.NewTimer(duration)                                                  //start timer.

	select { //used for selecting channel for given event.
	case enve := <-serv.ServerInfo.Inbox():
		//timer.Reset(duration)
		if enve.MsgId == 1 { //reply recvd

			var reply Reply
			err := json.Unmarshal([]byte(enve.Msg.(string)), &reply)
			if err != nil {
				if debug {
					log.Println("In Candidate, Unknown thing happened", enve.Msg.(string))
				}
			} else {
				//log.Println("Candidate : Serverid-", serv.ServerInfo.MyPid,"Leader is:-" ,RType.Leader,"Reply Recvd:-", reply, enve, enve.Msg.(string))
				if serv.Cur_Term() == reply.Term {
					//reply for current term recvd.
					if reply.Result {
						//true reply recvd.
						serv.ServState.followers[enve.Pid] = enve.Pid
						totalVotes := (len(serv.ServState.followers) + 1)
						if debug {
							log.Println("For Candidate : ", serv.ServerInfo.MyPid, "Vote Recvd from :-", enve.Pid, " total votes:-", totalVotes)
						}
						n := int((len(serv.ServerInfo.PeerIds) + 1) / 2)
						if n < totalVotes {
							//quorum.
							if debug {
								log.Println("Leader Declared:-", serv.ServerInfo.Pid(), "For Term:-", serv.Cur_Term())
							}
							//become leader.
							timer.Stop() //stop timer.
							mutex.Lock()
							//RType.Leader = serv.ServerInfo.Pid()
							serv.ServState.UpdateState(2) //update state to Leader.
							mutex.Unlock()

							// broadcast as a Leader.

							x := &Request{Term: serv.Cur_Term(), CandidateId: serv.ServerInfo.Pid(), LastLogIndex: serv.ServState.Log[int64(len(serv.ServState.Log))].Index, LastLogTerm: serv.ServState.Log[int64(len(serv.ServState.Log))].Term}
							data, err := json.Marshal(x)
							if err != nil {
								if debug {
									log.Println("Marshaling error", x)
								}
							}
							// braodcast the requestFor vote.
							serv.ServerInfo.Outbox() <- &cluster.Envelope{Pid: cluster.BROADCAST, MsgId: HEART, Msg: string(data)}

							return
						}
					} else {
						delete(serv.ServState.followers, enve.Pid)
						if debug {
							log.Println("Candidate : Serverid-", serv.ServerInfo.MyPid, "Rejection Recvd from :-", enve.Pid, "total votes:-", (len(serv.ServState.followers) + 1))
						}
					}
				} else {
					if debug {
						log.Println("Response ignored by Cand:-", serv.ServerInfo.MyPid, "For Term:", reply.Term)
					}
				}

			}
		} else if enve.MsgId == REQ {

			var req Request
			_ = json.Unmarshal([]byte(enve.Msg.(string)), &req) //decode message into Envelope object.
			//request received.
			//fmt.Println("Candidate : Serverid-", serv.ServerInfo.MyPid, "Request Recvd:-", req, enve)
			var reply *Reply
			/*	if req.CandidateId == RType.Leader && serv.Cur_Term() == req.Term {
					//Leader has send the message.
					reply = &Reply{Term: req.Term, Result: true}
					data, err := json.Marshal(reply)
					if err != nil {
						if debug{
							log.Println("In candidate receiving msgs: Marshaling error: ", reply)
						}
					}
					mutex.Lock()
					serv.ServState.UpdateVote_For(req.CandidateId)
					serv.ServState.UpdateState(0) //become follower.
					//serv.ServState.UpdateTerm(req.Term)
					serv.ServState.followers = make(map[int]int) //clear followers list.
					mutex.Unlock()
					envelope := cluster.Envelope{Pid: enve.Pid, MsgId: 1, Msg: string(data)}
					serv.ServerInfo.Outbox() <- &envelope
					if debug{
						log.Println("Cand -", serv.ServerInfo.MyPid, " has sent reply to Leader:-", enve.Pid)
					}
					timer.Stop() //stop timer.
					return
				} else if req.Term >= serv.Cur_Term()
			*/
			if req.LastLogIndex >= serv.ServState.LastApplied && req.Term > serv.Cur_Term() && req.LastLogTerm >= serv.ServState.Log[int64(len(serv.ServState.Log))].Term {
				// getting higher term.
				if debug {
					log.Println("Higher Term Recvd for Candidate ", serv.ServerInfo.MyPid, "from", enve.Pid)
				}
				reply = &Reply{Term: req.Term, Result: true}
				data, err := json.Marshal(reply)
				if err != nil {
					if debug {
						log.Println("In candidate receiving msgs: Marshaling error: ", reply)
					}
				}
				timer.Stop() //stop timer.
				mutex.Lock()
				serv.ServState.UpdateState(0) //become follower.
				serv.ServState.UpdateVote_For(req.CandidateId)
				serv.ServState.UpdateTerm(req.Term)
				serv.ServState.followers = make(map[int]int) //clear followers list.
				mutex.Unlock()
				envelope := cluster.Envelope{Pid: enve.Pid, MsgId: 1, Msg: string(data)}
				serv.ServerInfo.Outbox() <- &envelope
				//fmt.Println("from:-", serv.ServerInfo.MyPid, "to", enve.Pid, envelope)
				return
			} else {
				reply = &Reply{Term: req.Term, Result: false}
				data, err := json.Marshal(reply)
				if err != nil {
					if debug {
						log.Println("In candidate receiving msgs: Marshaling error: ", reply)
					}
				}
				envelope := cluster.Envelope{Pid: enve.Pid, MsgId: 1, Msg: string(data)}
				serv.ServerInfo.Outbox() <- &envelope
				if debug {
					log.Println("Request rcvd from -", enve.Pid, "to Cand(", serv.Cur_Term(), ") -", serv.ServerInfo.MyPid, "for Lower or equal Term:", reply.Term)
				}
				/*
					//now send request message and becareful about not to update term.

					ok := serv.ServState.UpdateState(1) // update state to be a candidate.
					x := &Request{Term: serv.Cur_Term(), CandidateId: serv.ServerInfo.Pid(),LastLogIndex:serv.ServState.Log[int64(len(serv.ServState.Log))].Index,LastLogTerm:serv.ServState.Log[int64(len(serv.ServState.Log))].Term}
					data, err = json.Marshal(x)
					//fmt.Println("data is:-", x)
					if err != nil {
						if debug{
							log.Println("Marshaling error", x)
						}
					}
					// braodcast the requestFor vote.
					//serv.ServerInfo.Outbox() <- &cluster.Envelope{Pid: cluster.BROADCAST, MsgId: 0, Msg: string(data)}
					serv.ServerInfo.Outbox() <- &cluster.Envelope{Pid: enve.Pid, MsgId: 0, Msg: string(data)}
					if !ok {
						println("error in updating state")
					}
				*/
			}

		} else if enve.MsgId == HEART { //HeartBeat rcvd.
			timer.Reset(duration) //reset timer
			var hrt HeartBeat
			err := json.Unmarshal([]byte(enve.Msg.(string)), &hrt) //decode message into Envelope object.
			if err != nil {                                        //error into parsing/decoding
				if debug {
					log.Println("Candidate,HeartBeat: Unmarshaling error:-\t", err)
				}
			}
			if debug {
				log.Println("Candidate:-Heartbeat Recvd for ", "sid-", serv.ServerInfo.MyPid, "from", enve.Pid, hrt)
			}
			// send appropriate reply.
			var reply *Reply
			if hrt.Term >= serv.Cur_Term() { //check whether leader is new enough.
				reply = &Reply{Term: hrt.Term, Result: true}
			} else {
				reply = &Reply{Term: hrt.Term, Result: false}
			}
			t_data, err := json.Marshal(reply)
			if err != nil {
				if debug {
					log.Println("Follower,HeartBeat:- Marshaling error: ", err)
				}
			}
			data := string(t_data)
			envelope := cluster.Envelope{Pid: enve.Pid, MsgId: 1, Msg: data}
			serv.ServerInfo.Outbox() <- &envelope
		}

	case <-timer.C:
		if debug {
			log.Println("Election Timer Timeout for:-", serv.ServerInfo.MyPid)
		}
		mutex.Lock()
		//serv.ServState.UpdateVote_For(serv.ServerInfo.Pid()) //giving him self vote.
		serv.ServState.UpdateVote_For(0)             //set vote to no-one.
		serv.ServState.UpdateState(0)                //become follower.
		serv.ServState.followers = make(map[int]int) //clear followers list.
		mutex.Unlock()
		timer.Stop()
		return
		/*
			ok := serv.ServState.UpdateState(1)                  // update state to be a candidate.
			if !ok {
				println("error in updating state")
			}*/
		//serv.ServState.UpdateTerm(serv.Cur_Term() + 1) //increment term by one.
		//x := &Request{Term: serv.Cur_Term(), CandidateId: serv.ServerInfo.Pid(),LastLogIndex:serv.ServState.Log[int64(len(serv.ServState.Log))].Index,LastLogTerm:serv.ServState.Log[int64(len(serv.ServState.Log))].Term}
		//data, err := json.Marshal(x)
		//fmt.Println("data is:-", x)
		//if err != nil {
		//	if debug{
		//		log.Println("Marshaling error", x)
		//	}
		//}
		// braodcast the requestFor vote.
		//serv.ServerInfo.Outbox() <- &cluster.Envelope{Pid: cluster.BROADCAST, MsgId: 0, Msg: string(data)}
		//timer = time.NewTimer(duration) //start timer.

	}
	timer.Stop()

}

//Server goes to Leader state.
func (serv *Server) StateLeader(mutex *sync.Mutex) {

	//duration := 1*time.Second + time.Duration(rand.Intn(51))*time.Millisecond//heartbeat timer.
	duration := 80*time.Millisecond + time.Duration(rand.Intn(51))*time.Millisecond //heartbeat time-duration.
	timer := time.NewTimer(duration)                                                //start timer.
	new_duration := 1*time.Second + duration
	timer_alive := time.NewTimer(new_duration) //start timer.
	for {
		select { //used for selecting channel for given event.
		case enve := <-serv.ServerInfo.Inbox():
			//timer.Reset(duration)
			if debug {
				log.Println("=====Reset Timer")
			}
			timer_alive.Reset(new_duration)
			switch enve.MsgId {
			case REP:
				//reply recvd
				var reply Reply
				err := json.Unmarshal([]byte(enve.Msg.(string)), &reply)
				if err != nil {
					if debug {
						log.Println("In Leader, Unknown thing happened", enve.Msg.(string))
					}
					timer.Stop()       //stop timer.
					timer_alive.Stop() //stop timer.
					return
				} else if reply.Term == serv.Cur_Term() {
					//fmt.Println("Leader : Serverid-", serv.ServerInfo.MyPid, "Reply Recvd:-", reply, enve, enve.Msg.(string))
					if reply.Result { //confirmation received
						//timer.Reset(duration)
						serv.ServState.followers[enve.Pid] = enve.Pid
						if serv.ServState.NextIndex[enve.Pid] == 0 && serv.ServState.MatchIndex[enve.Pid] == 0 {
							serv.ServState.NextIndex[enve.Pid] = serv.ServState.LastApplied
							serv.ServState.MatchIndex[enve.Pid] = 0
						}
						if debug {
							log.Println("Leader :", serv.ServerInfo.MyPid, "has received confirmation from", enve.Pid, "for term", reply.Term, "and total votes:-", (len(serv.ServState.followers) + 1))
						}
						//fmt.Println("For",serv.ServerInfo.MyPid,"Confirmation received from",enve.Pid,"total count:-",len(serv.ServState.followers))
					} else {
						if debug {
							log.Println("Leader :", serv.ServerInfo.MyPid, "has received rejection from", enve.Pid, "for term", reply.Term, "and total votes:-", (len(serv.ServState.followers) + 1))
						}
						delete(serv.ServState.followers, enve.Pid)

						totalVotes := (len(serv.ServState.followers) + 1)
						n := int((len(serv.ServerInfo.PeerIds) + 1) / 2)
						if n >= totalVotes {
							//RType.Leader = 0
							timer.Stop()       //stop timer.
							timer_alive.Stop() //stop timer.
							mutex.Lock()
							serv.ServState.UpdateVote_For(0)
							serv.ServState.UpdateState(0)                //become follower.
							serv.ServState.followers = make(map[int]int) //clear followers list.
							mutex.Unlock()
							return
						}
					}
				} else {
					if debug {
						log.Println("Leader :", serv.ServerInfo.MyPid, "has ignored reply from", enve.Pid, "for term", reply.Term, "and total votes:-", (len(serv.ServState.followers) + 1), " and Reply was", reply.Result)
					}
				}
				break
			case REQ:
				// Request Rcvd.
				var req Request
				_ = json.Unmarshal([]byte(enve.Msg.(string)), &req) //decode message into Envelope object.
				//request received.
				//fmt.Println("Leader : Serverid-", serv.ServerInfo.MyPid, "Request Recvd:-", req, enve)
				var reply *Reply
				if req.LastLogIndex >= serv.ServState.LastApplied && req.Term > serv.Cur_Term() && req.LastLogTerm >= serv.ServState.Log[int64(len(serv.ServState.Log))].Term {
					// getting higher term and it has not voted before.
					//RType.Leader = 0 //reset leader.
					if debug {
						log.Println("Leader(with Term", serv.Cur_Term(), ") :", serv.ServerInfo.MyPid, "has received higher term", req.Term, "from", enve.Pid)
					}
					reply = &Reply{Term: req.Term, Result: true}
					data, err := json.Marshal(reply)
					if err != nil {
						if debug {
							log.Println("In Leader receiving msgs: Marshaling error: ", reply)
						}
					}
					timer.Stop()       //stop timer.
					timer_alive.Stop() //stop timer.
					mutex.Lock()
					serv.ServState.UpdateState(0) //become follower.
					serv.ServState.UpdateVote_For(req.CandidateId)
					serv.ServState.UpdateTerm(req.Term)
					serv.ServState.followers = make(map[int]int)    //clear followers list.
					serv.ServState.NextIndex = make(map[int]int64)  //clear followers list.
					serv.ServState.MatchIndex = make(map[int]int64) //clear followers list.
					mutex.Unlock()
					envelope := cluster.Envelope{Pid: enve.Pid, MsgId: 1, Msg: string(data)}
					serv.ServerInfo.Outbox() <- &envelope
					//fmt.Println("from:-", serv.ServerInfo.MyPid, "to", enve.Pid, envelope)
					//serv.ServState.UpdateVote_For(enve.Pid)
					return
				} else {
					reply = &Reply{Term: req.Term, Result: false}
					data, err := json.Marshal(reply)
					if err != nil {
						if debug {
							log.Println("In Leader receiving msgs: Marshaling error: ", reply)
						}
					}
					envelope := cluster.Envelope{Pid: enve.Pid, MsgId: 1, Msg: string(data)}
					serv.ServerInfo.Outbox() <- &envelope
					if debug {
						log.Println("Leader(", serv.Cur_Term(), ") :", serv.ServerInfo.MyPid, "has received lesser or equal term req from ", enve.Pid, req.Term)
					}
					//now send request message and becareful about not to update term.

					x := &Request{Term: serv.Cur_Term(), CandidateId: serv.ServerInfo.Pid(), LastLogIndex: serv.ServState.Log[int64(len(serv.ServState.Log))].Index, LastLogTerm: serv.ServState.Log[int64(len(serv.ServState.Log))].Term}
					data, err = json.Marshal(x)
					//fmt.Println("data is:-", x)
					if err != nil {
						if debug {
							log.Println("Marshaling error", x)
						}
					}
					// braodcast the requestFor vote.
					serv.ServerInfo.Outbox() <- &cluster.Envelope{Pid: cluster.BROADCAST, MsgId: 0, Msg: string(data)}
				}

				break

			case AER:
				//timer.Reset(duration)
				//timer.Stop()
				serv.ServState.followers[enve.Pid] = enve.Pid
				var aer AE_Reply
				err := json.Unmarshal([]byte(enve.Msg.(string)), &aer) //decode message into Envelope object.
				if err != nil {                                        //error into parsing/decoding
					if debug {
						log.Println("Follower,AppendEntries: Unmarshaling error:-\t", err)
					}
				}
				if debug {
					log.Println("Leader:- Response for AppendEntries received for sid:-", serv.ServerInfo.MyPid, "from", enve.Pid, "data is:-", aer, "PrevLogIndex:-", aer.PrevLogIndex)
				}
				serv.ServState.NextIndex[enve.Pid] = aer.ExpectedIndex

				if aer.Success {
					serv.ServState.MatchIndex[enve.Pid] = aer.PrevLogIndex + 1
					total := 1
					n := int((len(serv.ServerInfo.PeerIds) + 1) / 2)
					for val, index := range serv.ServState.MatchIndex {
						if debug {
							log.Println("In Leader aer.Success:-", val, index, "commitIndex:-", serv.ServState.CommitIndex)
						}
						if index > serv.ServState.CommitIndex {
							total++
							if n < total {
								serv.ServState.CommitIndex++
								//t_data, err := json.Marshal(serv.ServState.CommitIndex)
								t_data := make([]byte, 8)
								binary.PutVarint(t_data, serv.ServState.CommitIndex)
								t_data1, err := json.Marshal(serv.ServState.Log[serv.ServState.CommitIndex])
								if err != nil {
									if debug {
										log.Println("In Leader :AER error in marshaling before writing to db:-", err)
									}
								}
								err = db.Put(t_data, t_data1, nil) //writting to database.
								/* //This helped to find one particular problem occured due to copying above code from follower's method and forgot one place to change it.
								if serv.ServState.CommitIndex == 16{
									log.Println("====================///////////////////////============================== serv.ServState.CommitIndex:-",serv.ServState.CommitIndex,err)
								}
								*/
								if err != nil {
									if debug {
										log.Println("In Leader:AER writing to db error.:-", err)
									}
								}
								temp := serv.ServState.CommitIndex
								serv.out <- &temp
								break
							}
						}
					}
				}
				/* //all new messages will be sent on timeout.
				_, exist := serv.ServState.Log[aer.ExpectedIndex]
				if exist {
					//entry := LogItem{Index: aer.ExpectedIndex, Term: int64(serv.ServState.my_term), Data: serv.ServState.Log[aer.ExpectedIndex]}
					entry := serv.ServState.Log[aer.ExpectedIndex]
					app := &AppendEntries{Term: serv.Cur_Term(), LeaderId: serv.ServerInfo.Pid(), PrevLogIndex: serv.ServState.Log[aer.ExpectedIndex-1].Index, PrevLogTerm: serv.ServState.Log[aer.ExpectedIndex-1].Term, Entries: entry, LeaderCommitIndex: serv.ServState.CommitIndex}
					data, err := json.Marshal(app)
					if err != nil {
						if debug {
							log.Println("Append Entry,Leader, Marshaling error", app, err)
						}
					} else {
						serv.ServerInfo.Outbox() <- &cluster.Envelope{Pid: enve.Pid, MsgId: APP, Msg: string(data)}
					}
					if debug {
						log.Println("Leader:- new log entry sent is:-", app, "NextIndex:-", serv.ServState.NextIndex, "MatchIndex", serv.ServState.MatchIndex)
					}
				}
				*/
				//timer.Reset(duration)
				break
				/*
					case L_I_REQ:
						log.Println("Leader:- L_I_REQ Request from client received", enve)
						break
					default:
						log.Println("Leader:- default", enve)
						break
				*/
			}

		case logitem := <-serv.in:
			log.Println("Leader:-logitem received from kv", logitem, "Log:-", serv.ServState.Log)
			/*

			*/

		case <-timer_alive.C:
			log.Println("===== Timer Expire")
			timer.Stop()       //stop timer.
			timer_alive.Stop() //stop timer.
			mutex.Lock()
			serv.ServState.UpdateState(0) //become follower.
			serv.ServState.UpdateVote_For(0)
			serv.ServState.followers = make(map[int]int)    //clear followers list.
			serv.ServState.NextIndex = make(map[int]int64)  //clear followers list.
			serv.ServState.MatchIndex = make(map[int]int64) //clear followers list.
			mutex.Unlock()
			return

		case <-timer.C:
			timer.Reset(duration)
			//send heartbeat to all servers
			x := &HeartBeat{Term: serv.Cur_Term(), LeaderId: serv.ServerInfo.Pid()}
			data, err := json.Marshal(x)
			if debug {
				log.Println("Timeout for Leader:-")
				serv.PrintData()
			}
			if err != nil {
				if debug {
					log.Println("Marshaling error", x, err)
				}
			}
			// braodcast the requestFor vote.

			//serv.ServerInfo.Outbox() <- &cluster.Envelope{Pid: cluster.BROADCAST, MsgId: HEART, Msg: string(data)}

			//send Append entry requests to followers
			for _, pid := range serv.ServState.followers {
				sent := false
				//log.Println("Inloop Timeout for Leader:-", serv.ServerInfo.Pid(), "ServerState is:-", serv.ServState)
				temp_pid := serv.ServState.NextIndex[pid]
				if temp_pid == 0 {
					temp_pid = serv.ServState.LastApplied
				}
				_, exist := serv.ServState.Log[temp_pid]
				if exist {
					if debug {
						log.Println("Leader, temp_pid:-", temp_pid, "serv.ServState.Log[temp_pid-1]", serv.ServState.Log[temp_pid-1])
					}
					//entry := LogItem{Index: serv.ServState.NextIndex[pid], Term: int64(serv.ServState.my_term), Data: serv.ServState.Log[serv.ServState.NextIndex[pid]]}
					entry := serv.ServState.Log[temp_pid]
					app := &AppendEntries{Term: serv.Cur_Term(), LeaderId: serv.ServerInfo.Pid(), PrevLogIndex: serv.ServState.Log[temp_pid-1].Index, PrevLogTerm: serv.ServState.Log[temp_pid-1].Term, Entries: entry, LeaderCommitIndex: serv.ServState.CommitIndex}
					data, err = json.Marshal(app)
					if err != nil {
						if debug {
							log.Println("Append Entry,Leader, Marshaling error", app, err)
						}
					} else {
						serv.ServerInfo.Outbox() <- &cluster.Envelope{Pid: pid, MsgId: APP, Msg: string(data)}
						sent = true
						if debug {
							log.Println("In timeout ,Leader:- new log entry sent is:-", app)
						}
					}
				}

				if !sent {
					serv.ServerInfo.Outbox() <- &cluster.Envelope{Pid: pid, MsgId: HEART, Msg: string(data)}
					if debug {
						log.Println("In timeout ,Leader:- heartbeat sent:-")
					}
				}
			}

			serv.ServState.followers = make(map[int]int) //clear followers list.

			/*
				x := &Request{Term: serv.Cur_Term(), CandidateId: serv.ServerInfo.Pid(),LastLogIndex:serv.ServState.Log[int64(len(serv.ServState.Log))].Index,LastLogTerm:serv.ServState.Log[int64(len(serv.ServState.Log))].Term}
				data, err := json.Marshal(x)
				if debug {
					log.Println("Timeout for Leader:-", serv.ServerInfo.Pid(), "Term", serv.Cur_Term())
				}
				if err != nil {
					if debug {
						log.Println("Marshaling error", x)
					}
				}
				// braodcast the requestFor vote.
				serv.ServerInfo.Outbox() <- &cluster.Envelope{Pid: cluster.BROADCAST, MsgId: 0, Msg: string(data)}
			*/
			/*
				for _, pid := range serv.ServState.followers {
					serv.ServerInfo.Outbox() <- &cluster.Envelope{Pid: pid, MsgId: 0, Msg: string(data)}
				}
			*/
			//log.Println("Leader -", serv.ServerInfo.MyPid, " is going for sleep.")
			//time.Sleep(20 * time.Second)
			//time.Sleep(time.Duration((rand.Intn(350) + serv.ServerInfo.MyPid*400)) * time.Millisecond) //minimum will be 400ms.
			//log.Println("Leader -", serv.ServerInfo.MyPid, " has awaken from sleep.")

		}
	}
	if debug {
		log.Println("Leader ouside switch")
	}

}
