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
	valid,rafttype := Raft.InitServer(myid, "./config.json",true )

	if valid{
		wg.Add(1)
	}

	go printData(rafttype)

	wg.Wait()
	return
}

func printData(rafttype *Raft.RaftType){
	for{
		time.Sleep(5*time.Second)
		println(rafttype.CurTerm(),rafttype.Leader())
	}
}
