package main

import (
	"flag"
	Raft "github.com/mgkanani/raft"
	"sync"
	"time"
)

var total_servers = 7

func main() {
	flagid := flag.Int("pid", 1, "flag type is integer")
	flag.Parse()
	myid := *flagid
	wg := new(sync.WaitGroup)
	//ch := make(chan int)
	//valid := InitServer(myid, "./config.json",true , ch)
	valid, rafttype := Raft.InitServer(myid, "./config.json", false)

	if valid {
		wg.Add(1)
	}

	go printData(rafttype,myid)

	wg.Wait()
	return
}

func printData(rafttype *Raft.RaftType,myid int) {
	for {
		time.Sleep(3100 * time.Millisecond)
		println("[",myid,"]","Current Term is :-",rafttype.CurTerm(), "Current Leader is:-",rafttype.Leader())
	}
}
