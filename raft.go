package raft

import (
		"encoding/json"
	"fmt"
	cluster "github.com/mgkanani/cluster"
	rand "math/rand"
	"time"
)

type Raft interface {
	Term() int
	isLeader() bool
}

type RaftType struct {
	leader int
}

type Reply struct{
	Term int
	Result bool
}

type Request struct{
	Term int
	CandidateId int
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

func (serv *ServerState) UpdateTerm(term int) {
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
			fmt.Println("Follower : Serverid-",serv.ServerInfo.MyPid)
			//follower

		FOLLOW:
			select { //used for selecting channel for given event.
			//case enve := <-ch:
			case enve := <-serv.ServerInfo.Inbox():
				//fmt.Println("Follower : Serverid-",serv.ServerInfo.MyPid," Msg Rcvd:-",enve)
				var req Request;
				err := json.Unmarshal([]byte(enve.Msg.(string)), &req) //decode message into Envelope object.
		                if err != nil {                        //error into parsing/decoding
                		        fmt.Println("Follower: Unmarshaling error:-\t", err)
		                }
				fmt.Println("Follower : Serverid-",serv.ServerInfo.MyPid," Request is:-.",req)

				if enve.Pid == serv.Vote() {
					fmt.Println("Follower : Serverid-",serv.ServerInfo.MyPid," heartbeat received for")
				} else {
					fmt.Println("Follower : Serverid-",serv.ServerInfo.MyPid," Request Recieved.",req)
					var reply *Reply
					if req.Term >= serv.Term() && (serv.Vote() == 0||serv.Vote() == serv.ServerInfo.Pid() ){ 
						// getting higher term and it has not voted before.
						reply = &Reply{Term:req.Term,Result:true}
                                                t_data, err := json.Marshal(reply)
						if err!=nil{
	                                                fmt.Println("Follower:- getting higher term:- Marshaling error: ",err);
        	                                }
						data:=string(t_data)
                                                serv.ServState.UpdateVote_For(req.CandidateId)
                                                envelope:=cluster.Envelope{Pid: enve.Pid, Msg:data} 
                                                serv.ServerInfo.Outbox() <- &envelope
                                                fmt.Println("Follower : Serverid-",serv.ServerInfo.MyPid,"reply sent on higher term recpt",req,reply,envelope);
						//err= json.Unmarshal(t_data, &reply)
						//fmt.Println("UUUUUUUUUUUU",reply)
					}else{//getting request for vote
						reply = &Reply{Term:req.Term,Result:false}
                                                data, err := json.Marshal(reply)
						if err!=nil{
	                                                fmt.Println("Follower:- getting request for vote:- Marshaling error: ",err);
        	                                }
                                                serv.ServState.UpdateVote_For(req.CandidateId)
                                                envelope:=cluster.Envelope{Pid: enve.Pid, Msg:string(data)} 
                                                serv.ServerInfo.Outbox() <- &envelope
                                                fmt.Println("Follower : Serverid-",serv.ServerInfo.MyPid,"reply sent on req for vote",req,reply,string(data),envelope);
					}
				}

			case <-time.After(2000+time.Duration(rand.Intn(151)*20) * time.Millisecond):
				println(serv.ServerInfo.MyPid,"No heartbeat has been received\n")
				serv.ServState.UpdateVote_For(0) //leader may crashed.
				//wait for 150-300ms
				sleep_time := time.Duration(150 + rand.Intn(151))
				sleep_time *= time.Millisecond
				time.After(sleep_time) //sleeps for random time in between 150ms and 300 ms.
				//fmt.Println(" awaken")
				if serv.Vote() == 0 {  //still no candidate exist.
					serv.ServState.UpdateVote_For(serv.ServerInfo.MyPid)      //giving him self vote.
					_ = serv.ServState.UpdateState(1) // update state to be a candidate.
					serv.ServState.UpdateTerm(serv.Term() + 1)    //increment term by one.
					var req *Request;
					req=&Request{Term:serv.Term(),CandidateId:serv.ServerInfo.Pid()}
					t_data, err := json.Marshal(req)
					if err!=nil{
                                                fmt.Println("Follower:- After Awaking :- Marshaling error: ",err);
					}else{
					data:=string(t_data)
					// braodcast the requestFor vote.
					envelope:=cluster.Envelope{Pid: cluster.BROADCAST, Msg:data}
                                        serv.ServerInfo.Outbox() <- &envelope
                                        fmt.Println("Follower : Serverid-",serv.ServerInfo.MyPid,"After Awaking,req sent for vote",req,"actual sent",data,envelope,serv);
					}
					break FOLLOW;
				}
			}

		} else if serv.ServState.my_state == 2 {
			//leader

		} else {
			//candidate
			fmt.Println("Candidate : Serverid-",serv.ServerInfo.MyPid)
		   CAND:
			select { //used for selecting channel for given event.
			case enve := <-serv.ServerInfo.Inbox():
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
	                      			fmt.Println("Candidate : Serverid-",serv.ServerInfo.MyPid,"Reply Recvd:-",reply,enve,enve.Msg.(string))
						if reply.Result{
		                                        serv.ServState.followers[enve.Pid]=enve.Pid
	                      				fmt.Println("Candidate : Serverid-",serv.ServerInfo.MyPid,"Confirmation Recvd:-",reply,enve)
        		                                //fmt.Println("For",serv.ServerInfo.MyPid,"Confirmation received from",enve.Pid,"total count:-",len(serv.ServState.followers))
						}
					}
                                }else{
				//request received.
	                      		fmt.Println("Candidate : Serverid-",serv.ServerInfo.MyPid,"Request Recvd:-",req,enve)
					var reply *Reply
                                        if req.Term >= serv.Term() && serv.Vote() == serv.ServerInfo.Pid(){
                                                // getting higher term and it has not voted before.
						fmt.Println("For",serv.ServerInfo.MyPid,"higher term received from",enve.Pid)
                                                reply = &Reply{Term:req.Term,Result:true}
                                                data, err := json.Marshal(reply)
                                                if err!=nil{
                                                        fmt.Println("In candidate receiving msgs: Marshaling error: ",reply);
                                                }
                                                serv.ServState.UpdateVote_For(req.CandidateId)
                                                envelope:=cluster.Envelope{Pid: enve.Pid, Msg:string(data)}
                                                serv.ServerInfo.Outbox() <- &envelope
                                                fmt.Println("from:-", serv.ServerInfo.MyPid,"to",enve.Pid,envelope);
	                                        serv.ServState.UpdateVote_For(enve.Pid)
						break CAND;
                                        }else{
                                                reply = &Reply{Term:req.Term,Result:false}
                                                data, err := json.Marshal(reply)
                                                if err!=nil{
                                                        fmt.Println("In candidate receiving msgs: Marshaling error: ",reply);
                                                }
                                                serv.ServState.UpdateVote_For(req.CandidateId)
                                                envelope:=cluster.Envelope{Pid: enve.Pid, Msg:string(data)}
                                                serv.ServerInfo.Outbox() <- &envelope
                                                fmt.Println("from:-", serv.ServerInfo.MyPid,"to",enve.Pid,envelope);
                                        }


				}

			case <-time.After(2000+time.Duration(rand.Intn(151)*20) * time.Millisecond):
                                serv.ServState.UpdateVote_For(serv.ServerInfo.MyPid)      //giving him self vote.
                                 ok := serv.ServState.UpdateState(1) // update state to be a candidate.
                                        serv.ServState.UpdateTerm(serv.Term() + 1)    //increment term by one.
                                        x:=&Request{Term:serv.Term(),CandidateId:serv.ServerInfo.Pid()}
                                        data, err := json.Marshal(x)
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

