package raft

import (
	"encoding/json"
	//"fmt"
	cluster "github.com/mgkanani/cluster"
	"log"
	rand "math/rand"
	"time"

//	"sync"
)

type Raft interface {
	Term() int
	isLeader() bool
}

type RaftType struct {
	Leader int
}

var RType = &RaftType{Leader: 0}

type Reply struct {
	Term   int
	Result bool
}

type Request struct {
	Term        int
	CandidateId int
}

type ServerState struct {
	my_term  int //default will be zero
	vote_for int //value will be pid of leader.
	my_state int
	/*
		0	-- follower
		1	-- candidate
		2	-- leader
	*/
	followers map[int]int
}

type Server struct {
	ServState  ServerState
	ServerInfo cluster.ServerType
}

func (serv *ServerState) UpdateTerm(term int) {
	log.Println("Term :-", term)
	serv.my_term = term
}

func (serv *ServerState) UpdateVote_For(term int) {
	serv.vote_for = term
}

func (serv *ServerState) UpdateState(new_state int) bool {
	if new_state > 2 || new_state < 0 {
		return false
	} else {
		serv.my_state = new_state
		//fmt.Println("stateUpdated to:-",new_state,"for",serv.ServerInfo.MyPid,serv)
		return true
	}
}

func (serv Server) Term() int {
	return serv.ServState.my_term
}

func (serv Server) Vote() int {
	return serv.ServState.vote_for
}

func (raft RaftType) isLeader() bool {
	if raft.Leader > 0 {
		return true
		/*	}
			if serv.ServState.my_state == 2 {
				return true
		*/
	} else {
		return false
	}
}

func InitServer(pid int, file string) Server {
	serv := Server{}
	serv.ServerInfo = cluster.New(pid, file)
	serv.ServState.followers = make(map[int]int)
	return serv
}

func (serv Server) Start(RType *RaftType) {

	for {
		//fmt.Println("For out side ",serv)
		if serv.ServState.my_state == 0 {
			//fmt.Println("Follower : Serverid-", serv.ServerInfo.MyPid)
			//follower
			duration := 1*time.Second + time.Duration(rand.Intn(151))*time.Millisecond
			//duration := time.Duration((rand.Intn(50)+serv.ServerInfo.MyPid*50)*12)*time.Millisecond
			timer := time.NewTimer(duration)

		FOLLOW:
			select { //used for selecting channel for given event.
			//case enve := <-ch:
			case enve := <-serv.ServerInfo.Inbox():
				if enve.MsgId == 0 {
					//Request Rcvd.
					//fmt.Println("Follower : Serverid-",serv.ServerInfo.MyPid," Msg Rcvd:-",enve)
					var req Request
					err := json.Unmarshal([]byte(enve.Msg.(string)), &req) //decode message into Envelope object.
					if err != nil {                                        //error into parsing/decoding
						log.Println("Follower: Unmarshaling error:-\t", err)
					}
					//fmt.Println("Follower : Serverid-", serv.ServerInfo.MyPid, " Request is:-.", req)

					if enve.Pid == serv.Vote() && RType.isLeader() {
						timer.Reset(duration)
						log.Println("Heartbeat Recvd for ", "sid-", serv.ServerInfo.MyPid, "from", enve.Pid)
					} else {
						//fmt.Println("Follower : Serverid-", serv.ServerInfo.MyPid, " Request Recieved.", req)
						var reply *Reply
						//if (req.Term >= serv.Term() && serv.Vote() == 0 )|| serv.Vote() == serv.ServerInfo.Pid() {
						if (req.Term > serv.Term() || serv.Vote() == 0) || serv.Vote() == serv.ServerInfo.Pid() {
							// getting higher term and it has not voted before or same leader with.
							reply = &Reply{Term: req.Term, Result: true}
							t_data, err := json.Marshal(reply)
							if err != nil {
								log.Println("Follower:- getting higher term:- Marshaling error: ", err)
							}
							data := string(t_data)
							serv.ServState.UpdateTerm(req.Term)
							timer.Reset(duration) //reset timer
							serv.ServState.UpdateVote_For(req.CandidateId)
							envelope := cluster.Envelope{Pid: enve.Pid, MsgId: 1, Msg: data}
							serv.ServerInfo.Outbox() <- &envelope
							log.Println("Higher Term:", req.Term, "Recvd for sid-", serv.ServerInfo.MyPid)
							time.After(1000)
							//err= json.Unmarshal(t_data, &reply)
							//fmt.Println("UUUUUUUUUUUU",reply)
						} else { //getting request for vote
							reply = &Reply{Term: req.Term, Result: false}
							data, err := json.Marshal(reply)
							if err != nil {
								log.Println("Follower:- getting request for vote:- Marshaling error: ", err)
							}
							serv.ServState.UpdateVote_For(req.CandidateId)
							envelope := cluster.Envelope{Pid: enve.Pid, MsgId: 1, Msg: string(data)}
							serv.ServerInfo.Outbox() <- &envelope
							log.Println("Follower : Serverid-", serv.ServerInfo.MyPid, "Rejected for", req.CandidateId, "on Term:", req.Term)
							time.After(1000)
						}
					}
				} else { //reply recvd. drop it.

				}

			//case <-time.After(2000+time.Duration(rand.Intn(151)*20) * time.Millisecond):
			//case <-time.After(2000000000):
			case <-timer.C:

				log.Println("Timeout for:-", serv.ServerInfo.MyPid)
				//declare leader has gone.
				RType.Leader = 0
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
						log.Println("Follower:- After Awaking :- Marshaling error: ", err)
					} else {
						data := string(t_data)
						// braodcast the requestFor vote.
						envelope := cluster.Envelope{Pid: cluster.BROADCAST, MsgId: 0, Msg: data}
						serv.ServerInfo.Outbox() <- &envelope
						time.After(1000)
						//fmt.Println("Follower : Serverid-", serv.ServerInfo.MyPid, "After Awaking,req sent for vote", req, "actual sent", data, envelope, serv)
					}
					timer.Stop()
					break FOLLOW
				}
			}
			timer.Stop()
		} else if serv.ServState.my_state == 1 {
			//candidate
			duration := 1*time.Second + time.Duration(rand.Intn(151))*time.Millisecond
			timer := time.NewTimer(duration)

			//fmt.Println("Candidate : Serverid-", serv.ServerInfo.MyPid)
		CAND:
			select { //used for selecting channel for given event.
			case enve := <-serv.ServerInfo.Inbox():
				timer.Reset(duration)
				if enve.MsgId == 1 { //reply recvd

					var reply Reply
					err := json.Unmarshal([]byte(enve.Msg.(string)), &reply)
					if err != nil {
						log.Println("In Candidate, Unknown thing happened", enve.Msg.(string))
					} else {
						//log.Println("Candidate : Serverid-", serv.ServerInfo.MyPid,"Leader is:-" ,RType.Leader,"Reply Recvd:-", reply, enve, enve.Msg.(string))
						if serv.Term() == reply.Term {
							//reply for current term recvd.
							if reply.Result {
								//true reply recvd.
								serv.ServState.followers[enve.Pid] = enve.Pid
								log.Println("For Candidate : ", serv.ServerInfo.MyPid, "Vote Recvd from :-", enve.Pid, " total votes:-", (len(serv.ServState.followers) + 1))
								n := int(len(serv.ServerInfo.PeerIds) / 2)
								if n <= len(serv.ServState.followers) {
									//quorum.
									log.Println("Leader Declared:-", serv.ServerInfo.Pid())
									//become leader.
									RType.Leader = serv.ServerInfo.Pid()
									serv.ServState.UpdateState(2) //become follower.
									// broadcast as a Leader.

									x := &Request{Term: serv.Term(), CandidateId: serv.ServerInfo.Pid()}
									data, err := json.Marshal(x)
									if err != nil {
										log.Println("Marshaling error", x)
									}
									// braodcast the requestFor vote.
									serv.ServerInfo.Outbox() <- &cluster.Envelope{Pid: cluster.BROADCAST, MsgId: 0, Msg: string(data)}
									timer.Stop()
									time.After(1000)
									break CAND
								}
							} else {
								delete(serv.ServState.followers, enve.Pid)
								log.Println("Candidate : Serverid-", serv.ServerInfo.MyPid, "Rejection Recvd from :-", enve.Pid, "total votes:-", (len(serv.ServState.followers) + 1))
							}
						} else {
							log.Println("Response neglected by Cand:-", serv.ServerInfo.MyPid, "For Term:", reply.Term)
						}

					}
				} else {

					var req Request
					_ = json.Unmarshal([]byte(enve.Msg.(string)), &req) //decode message into Envelope object.
					//request received.
					//fmt.Println("Candidate : Serverid-", serv.ServerInfo.MyPid, "Request Recvd:-", req, enve)
					var reply *Reply
					if req.CandidateId == RType.Leader && serv.Term() == req.Term {
						//Leader has send the message.
						reply = &Reply{Term: req.Term, Result: true}
						data, err := json.Marshal(reply)
						if err != nil {
							log.Println("In candidate receiving msgs: Marshaling error: ", reply)
						}
						serv.ServState.UpdateVote_For(req.CandidateId)
						serv.ServState.UpdateState(0) //become follower.
						//serv.ServState.UpdateTerm(req.Term)
						serv.ServState.followers = make(map[int]int) //clear followers list.
						envelope := cluster.Envelope{Pid: enve.Pid, MsgId: 1, Msg: string(data)}
						serv.ServerInfo.Outbox() <- &envelope
						log.Println("Cand -", serv.ServerInfo.MyPid, " has sent reply to Leader:-", enve.Pid)
						timer.Stop()
						//time.After(1000)
						break CAND
					} else if req.Term > serv.Term() {
						// getting higher term.
						log.Println("Higher Term Recvd for", serv.ServerInfo.MyPid, "from", enve.Pid)
						reply = &Reply{Term: req.Term, Result: true}
						data, err := json.Marshal(reply)
						if err != nil {
							log.Println("In candidate receiving msgs: Marshaling error: ", reply)
						}
						serv.ServState.UpdateVote_For(req.CandidateId)
						serv.ServState.UpdateState(0) //become follower.
						serv.ServState.UpdateTerm(req.Term)
						serv.ServState.followers = make(map[int]int) //clear followers list.
						envelope := cluster.Envelope{Pid: enve.Pid, MsgId: 1, Msg: string(data)}
						serv.ServerInfo.Outbox() <- &envelope
						//fmt.Println("from:-", serv.ServerInfo.MyPid, "to", enve.Pid, envelope)
						timer.Stop()
						//time.After(1000)
						break CAND
					} else {
						reply = &Reply{Term: req.Term, Result: false}
						data, err := json.Marshal(reply)
						if err != nil {
							log.Println("In candidate receiving msgs: Marshaling error: ", reply)
						}
						envelope := cluster.Envelope{Pid: enve.Pid, MsgId: 1, Msg: string(data)}
						serv.ServerInfo.Outbox() <- &envelope
						log.Println("Request rcvd from -", enve.Pid, "to Cand(", serv.Term(), ") -", serv.ServerInfo.MyPid, "for Lower or equal Term:", reply.Term)
						time.After(20) //wait for 20 nanosec.
						//now send request message and becareful about not to update term.

						ok := serv.ServState.UpdateState(1) // update state to be a candidate.
						x := &Request{Term: serv.Term(), CandidateId: serv.ServerInfo.Pid()}
						data, err = json.Marshal(x)
						//fmt.Println("data is:-", x)
						if err != nil {
							log.Println("Marshaling error", x)
						}
						// braodcast the requestFor vote.
						serv.ServerInfo.Outbox() <- &cluster.Envelope{Pid: cluster.BROADCAST, MsgId: 0, Msg: string(data)}
						if !ok {
							println("error in updating state")
						}
						time.After(1000)
					}

				}

			case <-timer.C:
				serv.ServState.UpdateVote_For(serv.ServerInfo.Pid()) //giving him self vote.
				ok := serv.ServState.UpdateState(1)                  // update state to be a candidate.
				serv.ServState.UpdateTerm(serv.Term() + 1)           //increment term by one.
				x := &Request{Term: serv.Term(), CandidateId: serv.ServerInfo.Pid()}
				data, err := json.Marshal(x)
				//fmt.Println("data is:-", x)
				if err != nil {
					log.Println("Marshaling error", x)
				}
				// braodcast the requestFor vote.
				serv.ServerInfo.Outbox() <- &cluster.Envelope{Pid: cluster.BROADCAST, MsgId: 0, Msg: string(data)}
				time.After(1000)
				if !ok {
					println("error in updating state")
				}
			}
			timer.Stop()
		} else {
			//leader

			//duration := 1*time.Second + time.Duration(rand.Intn(51))*time.Millisecond//heartbeat timer.
			duration := time.Duration(rand.Intn(51)) * time.Millisecond //heartbeat time-duration.
			timer := time.NewTimer(duration)

			//fmt.Println("Leader : Serverid-", serv.ServerInfo.MyPid)
		LEADER:
			select { //used for selecting channel for given event.
			case enve := <-serv.ServerInfo.Inbox():
				timer.Reset(duration)
				if enve.MsgId == 1 { //reply recvd

					var reply Reply
					err := json.Unmarshal([]byte(enve.Msg.(string)), &reply)
					if err != nil {
						log.Println("In Leader, Unknown thing happened", enve.Msg.(string))
						return
					} else if reply.Term == serv.Term() {
						//fmt.Println("Leader : Serverid-", serv.ServerInfo.MyPid, "Reply Recvd:-", reply, enve, enve.Msg.(string))
						if reply.Result { //confirmation received
							serv.ServState.followers[enve.Pid] = enve.Pid
							log.Println("Leader :", serv.ServerInfo.MyPid, "has received confirmation from", enve.Pid, "for term", reply.Term, "and total votes:-", (len(serv.ServState.followers) + 1))
							//fmt.Println("For",serv.ServerInfo.MyPid,"Confirmation received from",enve.Pid,"total count:-",len(serv.ServState.followers))
						} else {
							log.Println("Leader :", serv.ServerInfo.MyPid, "has received rejection from", enve.Pid, "for term", reply.Term, "and total votes:-", (len(serv.ServState.followers) + 1))
							delete(serv.ServState.followers, enve.Pid)
							n := int(len(serv.ServerInfo.PeerIds) / 2)
							if n >= len(serv.ServState.followers) {
								RType.Leader = 0
								timer.Stop()
								serv.ServState.UpdateVote_For(0)
								serv.ServState.UpdateState(0)                //become follower.
								serv.ServState.followers = make(map[int]int) //clear followers list.
								break LEADER
							}
						}
					} else {
						log.Println("Leader :", serv.ServerInfo.MyPid, "has ignored reply from", enve.Pid, "for term", reply.Term, "and total votes:-", (len(serv.ServState.followers) + 1), " and Reply was", reply.Result)
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
						RType.Leader = 0 //reset leader.
						log.Println("Leader(", serv.Term(), ") :", serv.ServerInfo.MyPid, "has received higher term", req.Term, "from", enve.Pid)
						reply = &Reply{Term: req.Term, Result: true}
						data, err := json.Marshal(reply)
						if err != nil {
							log.Println("In Leader receiving msgs: Marshaling error: ", reply)
						}
						serv.ServState.UpdateVote_For(req.CandidateId)
						serv.ServState.UpdateState(0) //become follower.
						serv.ServState.UpdateTerm(req.Term)
						serv.ServState.followers = make(map[int]int) //clear followers list.
						envelope := cluster.Envelope{Pid: enve.Pid, MsgId: 1, Msg: string(data)}
						serv.ServerInfo.Outbox() <- &envelope
						//fmt.Println("from:-", serv.ServerInfo.MyPid, "to", enve.Pid, envelope)
						serv.ServState.UpdateVote_For(enve.Pid)
						timer.Stop()
						//time.After(1000)
						break LEADER
					} else {
						//time.After(1000)
						reply = &Reply{Term: req.Term, Result: false}
						data, err := json.Marshal(reply)
						if err != nil {
							log.Println("In Leader receiving msgs: Marshaling error: ", reply)
						}
						envelope := cluster.Envelope{Pid: enve.Pid, MsgId: 1, Msg: string(data)}
						serv.ServerInfo.Outbox() <- &envelope
						log.Println("Leader(", serv.Term(), ") :", serv.ServerInfo.MyPid, "has received lesser or equal term req from ", enve.Pid, req.Term)
						time.After(20) //wait for 20 nanosec.
						//now send request message and becareful about not to update term.

						x := &Request{Term: serv.Term(), CandidateId: serv.ServerInfo.Pid()}
						data, err = json.Marshal(x)
						//fmt.Println("data is:-", x)
						if err != nil {
							log.Println("Marshaling error", x)
						}
						// braodcast the requestFor vote.
						serv.ServerInfo.Outbox() <- &cluster.Envelope{Pid: cluster.BROADCAST, MsgId: 0, Msg: string(data)}
					}

				}

			case <-timer.C:
				//send heartbeat to all followers
				//serv.ServState.UpdateVote_For(serv.ServerInfo.Pid()) //giving him self vote.
				//ok := serv.ServState.UpdateState(2)                  // update state to be a candidate.
				//serv.ServState.UpdateTerm(serv.Term() + 1)           //increment term by one.
				x := &Request{Term: serv.Term(), CandidateId: serv.ServerInfo.Pid()}
				data, err := json.Marshal(x)
				log.Println("Timeout for Leader:-", serv.ServerInfo.Pid(), "Term", serv.Term())
				if err != nil {
					log.Println("Marshaling error", x)
				}
				// braodcast the requestFor vote.
				for _, pid := range serv.ServState.followers {
					serv.ServerInfo.Outbox() <- &cluster.Envelope{Pid: pid, MsgId: 0, Msg: string(data)}
				}

				log.Println("Leader -", serv.ServerInfo.MyPid, " is going for sleep.")
				time.Sleep(30 * time.Second)
				log.Println("Leader -", serv.ServerInfo.MyPid, " has awaken from sleep.")
				time.After(1000)

			}
			timer.Stop()
		}
	}

}
