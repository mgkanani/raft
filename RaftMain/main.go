package main

import (
	"flag"
	Raft "github.com/mgkanani/raft"
//	"sync"
)

var total_servers = 7

func main() {
	flagid := flag.Int("pid", 1, "flag type is integer")
	flag.Parse()
        myid := *flagid
	//ch := make(chan int)
	//valid := InitServer(myid, "./config.json",true , ch)
	_ = Raft.InitServer(myid, "./config.json",true )

//	wg := new(sync.WaitGroup)
//	wg.Add(1)
//	wg.Wait()
	return

}
