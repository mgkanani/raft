package main

import (
	"flag"
	Raft "github.com/mgkanani/raft"
	"log"
	"net"
	"net/rpc"
	"strconv"
	"sync"
	"time"
)

var total_servers = 7

var rafttype *Raft.RaftType

type Test struct{}

func (t *Test) Add(id int, reply *Raft.Request) error {
	reply.Term = rafttype.CurTerm()
	reply.CandidateId = rafttype.Leader()

	return nil
}

func main() {
	flagid := flag.Int("pid", 1, "flag type is integer")
	flag.Parse()
	myid := *flagid
	wg := new(sync.WaitGroup)
	var valid bool
	//ch := make(chan int)
	//valid := InitServer(myid, "./config.json",true , ch)
	valid, rafttype = Raft.InitServer(myid, "./config.json", false)

	if valid {
		wg.Add(1)

		tst := new(Test)
		rpc.Register(tst)
		listener, e := net.Listen("tcp", "127.0.0.1:2134"+strconv.Itoa(myid))
		if e != nil {
			log.Fatal("listen error:", e)
		}
		for {
			if conn, err := listener.Accept(); err != nil {
				log.Fatal("accept error: " + err.Error())
			} else {
				log.Printf("new connection established\n")
				go rpc.ServeConn(conn)
			}
		}

		//go printData(rafttype,myid)
		wg.Done()
	}
	wg.Wait()
	return
}

func printData(rafttype *Raft.RaftType, myid int) {
	for {
		time.Sleep(3100 * time.Millisecond)
		println("[", myid, "]", "Current Term is :-", rafttype.CurTerm(), "Current Leader is:-", rafttype.Leader())
	}
}
