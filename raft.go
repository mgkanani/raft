package raft

import (
		"encoding/json"
	/*
		zmq "github.com/pebbe/zmq4"*/
	/*	"io/ioutil"
		"os"
		"strconv"*/
	"fmt"
	cluster "github.com/mgkanani/cluster"
	rand "math/rand"
	"time"

//	"sync"
)

type Raft interface {
	Term() int
	isLeader() bool
}

type RaftType struct {
	leader int
}

type Reply struct{
	term int
	result bool
}

type Request struct{
	term int
	candidateId int
}

type ServerState struct{
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
	ServState ServerState
	ServerInfo cluster.ServerType
}

func (serv *Server) UpdateTerm(term int) {
	serv.ServState.my_term = term
}

func (serv *Server) UpdateVote_For(term int) {
	serv.ServState.vote_for = term
}

func (serv *Server) UpdateState(new_state int) bool {
	if new_state > 2 || new_state < 0 {
		return false
	} else {
		serv.ServState.my_state = new_state
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

func (serv Server) isLeader() bool {
	if serv.ServState.my_state == 2 {
		return true
	} else {
		return false
	}
}

func InitServer(pid int, file string) Server {
	serv := Server{}
	serv.ServerInfo = cluster.New(pid, file)
	return serv
}

func (serv Server) Start() {
	for {
		//fmt.Println("For out side ",serv)
		if serv.ServState.my_state == 0 {
			fmt.Println("Server-",serv.ServerInfo.MyPid," In Follower.")
			//follower

			//ch := make(chan *cluster.Envelope) //use for filtering messages that are coming from leader.
			//go serv.Filter(ch);
		FOLLOW:
			select { //used for selecting channel for given event.
			//case enve := <-ch:
			case enve := <-serv.ServerInfo.Inbox():
				var req Request;
				err := json.Unmarshal([]byte(enve.Msg.(string)), &req) //decode message into Envelope object.
		                if err != nil {                        //error into parsing/decoding
                		        fmt.Println("In Candidate Unmarshaling error:-\t", err,enve)
		                }

				//str := envelope.Msg.(string)
				//fmt.Println("At:-", serv.ServerInfo.MyPid, "\tReceived from:-", enve.Pid, "\tterm is:-", enve.MsgId, enve)
				fmt.Println("At:-", serv.ServerInfo.MyPid, "\tReceived from:-", enve.Pid, enve)

				if enve.Pid == serv.Vote() {
					fmt.Println("heartbeat received")
				} else {
					var reply *Reply
					if req.term >= serv.Term() && (serv.Vote() == 0||serv.Vote() == serv.ServerInfo.Pid() ){ 
						// getting higher term and it has not voted before.
						reply = &Reply{term:req.term,result:true}
                                                t_data, err := json.Marshal(*reply)
						if err!=nil{
	                                                fmt.Println("In candidate receiving msgs: Marshaling error: ",reply);
        	                                }
						data:=string(t_data)
                                                serv.UpdateVote_For(req.candidateId)
                                                envelope:=cluster.Envelope{Pid: enve.Pid, Msg:&data} 
                                                serv.ServerInfo.Outbox() <- &envelope
                                                fmt.Println("from:-", serv.ServerInfo.MyPid,"to",enve.Pid,envelope);
					}else{
						reply = &Reply{term:req.term,result:false}
                                                data, err := json.Marshal(*reply)
						if err!=nil{
	                                                fmt.Println("In candidate receiving msgs: Marshaling error: ",reply);
        	                                }
                                                serv.UpdateVote_For(req.candidateId)
                                                envelope:=cluster.Envelope{Pid: enve.Pid, Msg:string(data)} 
                                                serv.ServerInfo.Outbox() <- &envelope
                                                fmt.Println("from:-", serv.ServerInfo.MyPid,"to",enve.Pid,envelope);
					}
				}

			//case <-time.After(time.Duration(150+rand.Intn(151)) * time.Millisecond):
			case <-time.After(2000+time.Duration(rand.Intn(151)*20) * time.Millisecond):
				println(serv.ServerInfo.MyPid,"No heartbeat has been received\n")
				serv.UpdateVote_For(0) //leader may crashed.
				//wait for 150-300ms
				sleep_time := time.Duration(150 + rand.Intn(151))
				sleep_time *= time.Millisecond
				time.After(sleep_time) //sleeps for random time in between 150ms and 300 ms.
				//fmt.Println(" awaken")
				if serv.Vote() == 0 {  //still no candidate exist.
					serv.UpdateVote_For(serv.ServerInfo.MyPid)      //giving him self vote.
					ok := serv.UpdateState(1) // update state to be a candidate.
					serv.UpdateTerm(serv.Term() + 1)    //increment term by one.
					x:=&Request{term:serv.Term(),candidateId:serv.ServerInfo.Pid()}
					t_data, err := json.Marshal(*x)
					fmt.Println("data is:-",x);
					if err!=nil{
						fmt.Println("Marshaling error",x);
					}
					data:=string(t_data)
					// braodcast the requestFor vote.
					envelope:=cluster.Envelope{Pid: cluster.BROADCAST, Msg:string(data)}
                                        serv.ServerInfo.Outbox() <- &envelope
                                        fmt.Println("from:-", serv.ServerInfo.MyPid,"to",envelope.Pid,envelope);
					if !ok {
						println("error in updating state")
					}
					break FOLLOW;
				}
			}

		} else if serv.ServState.my_state == 2 {
			//leader

		} else {
			//candidate
			// go for election.
			//count=0
			fmt.Println("For",serv.ServerInfo.MyPid," In Candidate.")
		   CAND:
			select { //used for selecting channel for given event.

			case enve := <-serv.ServerInfo.Inbox():
                           //     fmt.Println("Reply At:-", serv.ServerInfo.MyPid, "\tReceived from:-", enve.Pid, "\tterm is:-", enve.MsgId,enve.Msg == "confirm",enve)



                                var req Request;
                                err := json.Unmarshal([]byte(enve.Msg.(string)), &req) //decode message into Envelope object.
                                if err != nil {                        //error into parsing/decoding
                                        //fmt.Println("In Candidate Unmarshaling error:-\t", err,enve)
					//It may be reply.
					var reply Reply
					err := json.Unmarshal([]byte(enve.Msg.(string)), &reply)
	                                if err != nil {
						fmt.Println("In Candidate, Unknown thing happened",enve.Msg.(string));
					}else{
						if reply.result{
		                                        serv.ServState.followers[enve.Pid]=enve.Pid
        		                                fmt.Println("For",serv.ServerInfo.MyPid,"Confirmation received from",enve.Pid,"total count:-",len(serv.ServState.followers))
						}
					}
                                }else{
				//request received.
					var reply *Reply
                                        if req.term >= serv.Term() && serv.Vote() == serv.ServerInfo.Pid(){
                                                // getting higher term and it has not voted before.
						fmt.Println("For",serv.ServerInfo.MyPid,"higher term received from",enve.Pid)
                                                reply = &Reply{term:req.term,result:true}
                                                data, err := json.Marshal(*reply)
                                                if err!=nil{
                                                        fmt.Println("In candidate receiving msgs: Marshaling error: ",reply);
                                                }
                                                serv.UpdateVote_For(req.candidateId)
                                                envelope:=cluster.Envelope{Pid: enve.Pid, Msg:string(data)}
                                                serv.ServerInfo.Outbox() <- &envelope
                                                fmt.Println("from:-", serv.ServerInfo.MyPid,"to",enve.Pid,envelope);
	                                        serv.UpdateVote_For(enve.Pid)
						break CAND;
                                        }else{
                                                reply = &Reply{term:req.term,result:false}
                                                data, err := json.Marshal(*reply)
                                                if err!=nil{
                                                        fmt.Println("In candidate receiving msgs: Marshaling error: ",reply);
                                                }
                                                serv.UpdateVote_For(req.candidateId)
                                                envelope:=cluster.Envelope{Pid: enve.Pid, Msg:string(data)}
                                                serv.ServerInfo.Outbox() <- &envelope
                                                fmt.Println("from:-", serv.ServerInfo.MyPid,"to",enve.Pid,envelope);
                                        }


				}

			case <-time.After(2000+time.Duration(rand.Intn(151)*20) * time.Millisecond):

/*
				serv.UpdateVote_For(serv.ServerInfo.MyPid)      //giving him self vote.
				ok := serv.UpdateState(1) // update state to be a candidate.
				msg := "hi"
				serv.UpdateTerm(serv.Term() + 1)    //increment term by one.
				// braodcast the requestFor vote.
				serv.ServerInfo.Outbox() <- &cluster.Envelope{Pid: cluster.BROADCAST, MsgId: int64(serv.Term()), Msg: string(msg)} 
*/
                                serv.UpdateVote_For(serv.ServerInfo.MyPid)      //giving him self vote.
                                 ok := serv.UpdateState(1) // update state to be a candidate.
                                        serv.UpdateTerm(serv.Term() + 1)    //increment term by one.
                                        x:=&Request{term:serv.Term(),candidateId:serv.ServerInfo.Pid()}
                                        data, err := json.Marshal(*x)
                                        fmt.Println("data is:-",x);
                                        if err!=nil{
                                                fmt.Println("Marshaling error",x);
                                        }
                                        // braodcast the requestFor vote.
                                        serv.ServerInfo.Outbox() <- &cluster.Envelope{Pid: cluster.BROADCAST, Msg: string(data)}
				if !ok {
					println("error in updating state")
				}
			}

		}
	}

}

func (serv Server) Filter(enve chan *cluster.Envelope) {
	//var mut sync.RWMutex
	for {
		if serv.ServState.my_state == 0 { //ensuring it is in follower state
			//mut.Lock()
			y := <-serv.ServerInfo.Inbox()
			x := y
			enve <- x
			fmt.Println("In Filter:- :", x.Pid, serv.ServState.vote_for, x.Pid == 2, "\t", x)
			/*		//if x.Pid==serv.vote_for {
					if x.Pid==2 {
					//heartbeat or message from leader recieved.
					//fmt.Println("In Filter:- :",x.Pid,serv.vote_for,x.Pid==2,"\t",x)
					enve<-x;
					fmt.Println("In Filter:- :",x.Pid,serv.vote_for,x.Pid==2,"\t",x)
					}
			*/
			//mut.Unlock()
		} else {
			break
		}
	}
}
