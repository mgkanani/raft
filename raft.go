package raft

import (
	"encoding/json"
	cluster "github.com/mgkanani/cluster"
	"log"
	rand "math/rand"
	"time"
//	"fmt"
)

type RaftType struct{
	serv *Server;
}

type Raft interface {
	CurTerm() int      //returns the current term number.
	Leader() int   //returns id of a leader if there exist, otherwise returns -1;
}

//Msg Type,whether it is request or reply.
type MsgType struct {
	MType byte        // 1 for reply, 0 for request.
	Msg   interface{} //actual object/message.
}

type Reply struct {
	//Reply strucrure.
	Term   int
	Result bool // true if follower has voted, otherwise false for rejection for vote.
}

type Request struct {
	//Request message structure.
	Term        int //Request for vote for this Term.
	CandidateId int //Requested candidate-id.
}

// Server State data structure
type ServerState struct {
	my_term  int //default will be zero
	vote_for int //value will be pid of leader.
	my_state int
	followers map[int]int
}

const (
	FOLLOWER  = 0
	CANDIDATE = 1
	LEADER    = 2
)

var debug = true

type Server struct {
	ServState  ServerState        //Server-State informtion are stored.
	ServerInfo cluster.ServerType //Server meta information will be stored like ip,port.
}

//Update the Term
func (serv *ServerState) UpdateTerm(term int) {
	if debug {
		log.Println("Term :-", term)
	}
	serv.my_term = term
}

//Server update it's variable for which it has voted.
func (serv *ServerState) UpdateVote_For(term int) {
	serv.vote_for = term
}

//Server updates it's state and returns true if it is successful, otherwise false.
func (serv *ServerState) UpdateState(new_state int) bool {
	if new_state > 2 || new_state < 0 {
		return false
	} else {
		serv.my_state = new_state
		//fmt.Println("stateUpdated to:-",new_state,"for",serv.ServerInfo.MyPid,serv)
		return true
	}
}

//returns the current Term number.
func (serv Server) Term() int {
	return serv.ServState.my_term
}

//returns the Server-id for which server it has voted.
func (serv Server) Vote() int {
	return serv.ServState.vote_for
}

func (rt RaftType) CurTerm() int{
	return rt.serv.Term()
}


func (rt RaftType) Leader() int{
	if rt.serv.ServState.my_state == FOLLOWER{
		return rt.serv.ServState.vote_for
	}else if rt.serv.ServState.my_state == LEADER{
		return rt.serv.ServState.vote_for
	}
	return -1
}


func (rt *RaftType) setServer(serv *Server) {
	rt.serv=serv
}

//Initializes the servers with given parameters.
func InitServer(pid int, file string, dbg bool) (bool,*RaftType) {
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

	serv.ServerInfo = cluster.New(pid, file)
	serv.ServState.followers = make(map[int]int)

	if debug {
		log.Println(rtype, serv.ServerInfo.Valid,serv)
	}
	if serv.ServerInfo.Valid {
		go serv.start()
	}
	return serv.ServerInfo.Valid,&rtype
}

//starts the leader election process.
//func (serv Server) start(RType *RaftType, ch chan int)
func (serv *Server) start() {

	for {
		//fmt.Println("For out side ",serv)

		switch serv.ServState.my_state {
		case FOLLOWER:
			//follower
			//fmt.Println("Follower : Serverid-", serv.ServerInfo.MyPid)
			serv.StateFollower()
		case CANDIDATE:
			//candidate
			//fmt.Println("Candidate : Serverid-", serv.ServerInfo.MyPid)
			serv.StateCandidate()
		case LEADER:
			//leader
			//fmt.Println("Leader : Serverid-", serv.ServerInfo.MyPid)
			serv.StateLeader()
		}
	}

}

//Server goes to Follower state.
func (serv *Server) StateFollower() {
	duration := 1*time.Second + time.Duration(rand.Intn(151))*time.Millisecond
	//duration := 600*time.Millisecond
	//duration := time.Duration((rand.Intn(50)+serv.ServerInfo.MyPid*50)*12)*time.Millisecond
	timer := time.NewTimer(duration)

	select { //used for selecting channel for given event.
	case enve := <-serv.ServerInfo.Inbox():
		if enve.MsgId == 0 {
			//Request Rcvd.
			var req Request
			err := json.Unmarshal([]byte(enve.Msg.(string)), &req) //decode message into Envelope object.
			if err != nil {                                        //error into parsing/decoding
				if debug {
					log.Println("Follower: Unmarshaling error:-\t", err)
				}
			}
			//fmt.Println("Follower : Serverid-", serv.ServerInfo.MyPid, " Request is:-.", req)

			if req.Term < serv.Term() {
				//just drop the message or reject the request.

			} else if (enve.Pid == serv.Vote() && req.Term == serv.Term()) || req.Term > serv.Term() {
				//heartbeat received or higher term received, reset timer,send accept for request.
				timer.Reset(duration)
				//fmt.Println("Follower : Serverid-", serv.ServerInfo.MyPid, " Request Recieved.", req)
				var reply *Reply
				// getting higher term and it has not voted before or same leader with.
				reply = &Reply{Term: req.Term, Result: true}
				t_data, err := json.Marshal(reply)
				if err != nil {
					if debug {
						log.Println("Follower:- getting higher term:- Marshaling error: ", err)
					}
				}
				data := string(t_data)
				serv.ServState.UpdateTerm(req.Term)
				timer.Reset(duration) //reset timer
				serv.ServState.UpdateVote_For(req.CandidateId)
				envelope := cluster.Envelope{Pid: enve.Pid, MsgId: 1, Msg: data}
				serv.ServerInfo.Outbox() <- &envelope
				if enve.Pid == serv.Vote() {
					if debug {
						log.Println("Heartbeat Recvd for ", "sid-", serv.ServerInfo.MyPid, "from", enve.Pid)
					}
				} else {
					if debug {
						log.Println("Higher Term:", req.Term, "Recvd for Follower -", serv.ServerInfo.MyPid)
					}
				}
			} else { //getting request for vote
				reply := &Reply{Term: req.Term, Result: false}
				data, err := json.Marshal(reply)
				if err != nil {
					if debug {
						log.Println("Follower:- getting request for vote:- Marshaling error: ", err)
					}
				}
				serv.ServState.UpdateVote_For(req.CandidateId)
				envelope := cluster.Envelope{Pid: enve.Pid, MsgId: 1, Msg: string(data)}
				serv.ServerInfo.Outbox() <- &envelope
				if debug {
					log.Println("Follower : Serverid-", serv.ServerInfo.MyPid, "Rejected for", req.CandidateId, "on Term:", req.Term)
				}
			}
		} else { //reply recvd. drop it.
		}

	//case <-time.After(2000+time.Duration(rand.Intn(151)*20) * time.Millisecond):
	//case <-time.After(2000000000):
	case <-timer.C:

		if debug {
			log.Println("Timeout for:-", serv.ServerInfo.MyPid)
		}
		//declare leader has gone.
		//RType.Leader = 0
		serv.ServState.UpdateVote_For(0) //leader may crashed.
		//wait for 150-300ms
		sleep_time := time.Duration(150 + rand.Intn(151))
		sleep_time *= time.Millisecond
		time.After(sleep_time) //sleeps for random time in between 150ms and 300 ms.
		//fmt.Println(" awaken")
		if serv.Vote() == 0 { //still no candidate exist.
			serv.ServState.UpdateVote_For(serv.ServerInfo.MyPid) //giving him self vote.
			_ = serv.ServState.UpdateState(1)                    // update state to be a candidate.
			serv.ServState.UpdateTerm(serv.Term() + 1)           //increment term by one.
			var req *Request
			req = &Request{Term: serv.Term(), CandidateId: serv.ServerInfo.Pid()}
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
func (serv *Server) StateCandidate() {

	//duration := 1*time.Second + time.Duration(rand.Intn(151))*time.Millisecond
	//duration := 150*time.Millisecond + time.Duration(rand.Intn(151))*time.Millisecond //choose duration between 150-300ms.
	duration := 550*time.Millisecond + time.Duration(rand.Intn(151))*time.Millisecond //choose duration between 150-300ms.
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
				if serv.Term() == reply.Term {
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
								log.Println("Leader Declared:-", serv.ServerInfo.Pid(), "For Term:-", serv.Term())
							}
							//become leader.
							//RType.Leader = serv.ServerInfo.Pid()
							serv.ServState.UpdateState(2) //update state to Leader.
							// broadcast as a Leader.

							x := &Request{Term: serv.Term(), CandidateId: serv.ServerInfo.Pid()}
							data, err := json.Marshal(x)
							if err != nil {
								if debug {
									log.Println("Marshaling error", x)
								}
							}
							// braodcast the requestFor vote.
							serv.ServerInfo.Outbox() <- &cluster.Envelope{Pid: cluster.BROADCAST, MsgId: 0, Msg: string(data)}
							timer.Stop() //stop timer.
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
		} else {

			var req Request
			_ = json.Unmarshal([]byte(enve.Msg.(string)), &req) //decode message into Envelope object.
			//request received.
			//fmt.Println("Candidate : Serverid-", serv.ServerInfo.MyPid, "Request Recvd:-", req, enve)
			var reply *Reply
			/*	if req.CandidateId == RType.Leader && serv.Term() == req.Term {
					//Leader has send the message.
					reply = &Reply{Term: req.Term, Result: true}
					data, err := json.Marshal(reply)
					if err != nil {
						if debug{
							log.Println("In candidate receiving msgs: Marshaling error: ", reply)
						}
					}
					serv.ServState.UpdateVote_For(req.CandidateId)
					serv.ServState.UpdateState(0) //become follower.
					//serv.ServState.UpdateTerm(req.Term)
					serv.ServState.followers = make(map[int]int) //clear followers list.
					envelope := cluster.Envelope{Pid: enve.Pid, MsgId: 1, Msg: string(data)}
					serv.ServerInfo.Outbox() <- &envelope
					if debug{
						log.Println("Cand -", serv.ServerInfo.MyPid, " has sent reply to Leader:-", enve.Pid)
					}
					timer.Stop() //stop timer.
					return
				} else if req.Term >= serv.Term()
			*/
			if req.Term > serv.Term() {
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
				serv.ServState.UpdateVote_For(req.CandidateId)
				serv.ServState.UpdateState(0) //become follower.
				serv.ServState.UpdateTerm(req.Term)
				serv.ServState.followers = make(map[int]int) //clear followers list.
				envelope := cluster.Envelope{Pid: enve.Pid, MsgId: 1, Msg: string(data)}
				serv.ServerInfo.Outbox() <- &envelope
				//fmt.Println("from:-", serv.ServerInfo.MyPid, "to", enve.Pid, envelope)
				timer.Stop() //stop timer.
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
					log.Println("Request rcvd from -", enve.Pid, "to Cand(", serv.Term(), ") -", serv.ServerInfo.MyPid, "for Lower or equal Term:", reply.Term)
				}
				/*
					//now send request message and becareful about not to update term.

					ok := serv.ServState.UpdateState(1) // update state to be a candidate.
					x := &Request{Term: serv.Term(), CandidateId: serv.ServerInfo.Pid()}
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

		}

	case <-timer.C:
		if debug {
			log.Println("Election Timer Timeout for:-", serv.ServerInfo.MyPid)
		}
		//serv.ServState.UpdateVote_For(serv.ServerInfo.Pid()) //giving him self vote.
		serv.ServState.UpdateVote_For(0)             //set vote to no-one.
		serv.ServState.UpdateState(0)                //become follower.
		serv.ServState.followers = make(map[int]int) //clear followers list.
		timer.Stop()
		return
		/*
			ok := serv.ServState.UpdateState(1)                  // update state to be a candidate.
			if !ok {
				println("error in updating state")
			}*/
		//serv.ServState.UpdateTerm(serv.Term() + 1) //increment term by one.
		//x := &Request{Term: serv.Term(), CandidateId: serv.ServerInfo.Pid()}
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
func (serv *Server) StateLeader() {

	//duration := 1*time.Second + time.Duration(rand.Intn(51))*time.Millisecond//heartbeat timer.
	duration := time.Duration(rand.Intn(51)) * time.Millisecond //heartbeat time-duration.
	timer := time.NewTimer(duration)                            //start timer.

	select { //used for selecting channel for given event.
	case enve := <-serv.ServerInfo.Inbox():
		//timer.Reset(duration)
		if enve.MsgId == 1 { //reply recvd

			var reply Reply
			err := json.Unmarshal([]byte(enve.Msg.(string)), &reply)
			if err != nil {
				if debug {
					log.Println("In Leader, Unknown thing happened", enve.Msg.(string))
				}
				timer.Stop() //stop timer.
				return
			} else if reply.Term == serv.Term() {
				//fmt.Println("Leader : Serverid-", serv.ServerInfo.MyPid, "Reply Recvd:-", reply, enve, enve.Msg.(string))
				if reply.Result { //confirmation received
					serv.ServState.followers[enve.Pid] = enve.Pid
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
						serv.ServState.UpdateVote_For(0)
						serv.ServState.UpdateState(0)                //become follower.
						serv.ServState.followers = make(map[int]int) //clear followers list.
						timer.Stop()                                 //stop timer.
						return
					}
				}
			} else {
				if debug {
					log.Println("Leader :", serv.ServerInfo.MyPid, "has ignored reply from", enve.Pid, "for term", reply.Term, "and total votes:-", (len(serv.ServState.followers) + 1), " and Reply was", reply.Result)
				}
			}
		} else {
			// Request Rcvd.
			var req Request
			_ = json.Unmarshal([]byte(enve.Msg.(string)), &req) //decode message into Envelope object.
			//request received.
			//fmt.Println("Leader : Serverid-", serv.ServerInfo.MyPid, "Request Recvd:-", req, enve)
			var reply *Reply
			if req.Term > serv.Term() {
				// getting higher term and it has not voted before.
				//RType.Leader = 0 //reset leader.
				if debug {
					log.Println("Leader(with Term", serv.Term(), ") :", serv.ServerInfo.MyPid, "has received higher term", req.Term, "from", enve.Pid)
				}
				reply = &Reply{Term: req.Term, Result: true}
				data, err := json.Marshal(reply)
				if err != nil {
					if debug {
						log.Println("In Leader receiving msgs: Marshaling error: ", reply)
					}
				}
				serv.ServState.UpdateVote_For(req.CandidateId)
				serv.ServState.UpdateState(0) //become follower.
				serv.ServState.UpdateTerm(req.Term)
				serv.ServState.followers = make(map[int]int) //clear followers list.
				envelope := cluster.Envelope{Pid: enve.Pid, MsgId: 1, Msg: string(data)}
				serv.ServerInfo.Outbox() <- &envelope
				//fmt.Println("from:-", serv.ServerInfo.MyPid, "to", enve.Pid, envelope)
				serv.ServState.UpdateVote_For(enve.Pid)
				timer.Stop() //stop timer.
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
					log.Println("Leader(", serv.Term(), ") :", serv.ServerInfo.MyPid, "has received lesser or equal term req from ", enve.Pid, req.Term)
				}
				//now send request message and becareful about not to update term.

				x := &Request{Term: serv.Term(), CandidateId: serv.ServerInfo.Pid()}
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

		}

	case <-timer.C:
		//send heartbeat to all servers
		x := &Request{Term: serv.Term(), CandidateId: serv.ServerInfo.Pid()}
		data, err := json.Marshal(x)
		if debug {
			log.Println("Timeout for Leader:-", serv.ServerInfo.Pid(), "Term", serv.Term())
		}
		if err != nil {
			if debug {
				log.Println("Marshaling error", x)
			}
		}
		// braodcast the requestFor vote.
		serv.ServerInfo.Outbox() <- &cluster.Envelope{Pid: cluster.BROADCAST, MsgId: 0, Msg: string(data)}
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
	timer.Stop()

}
