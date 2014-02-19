package raft

import (
	/*	"flag"
		"bufio"
		"os"
		"fmt"
		Raft "github.com/mgkanani/raft"*/
	//  "io/ioutil"
	"bufio"
	"os"
	"sync"
	"testing"
	"log"
)

func TestRaft(t *testing.T) {
	/*        flagid := flag.Int("pid", 1, "flag type is integer")
	          //myid := flag.Int("pid",1,"flag type is integer")
	          flag.Parse()

	          myid := *flagid
	          server := Raft.InitServer(myid, "./config.json")
	          server.Start();

	          stdin := bufio.NewReader(os.Stdin)
	          fmt.Print("\nPress enter to start\n")
	          _, err := stdin.ReadString('\n')
	          if err != nil {
	                  fmt.Print(err)
	                  return
	          }
	*/
	raftType := &RaftType{Leader: 0}

	wg := new(sync.WaitGroup)
	for i := 1; i < 6; i++ {
		wg.Add(1)
		go StartServer(i, wg, raftType)
	}
	wg.Wait()
	return

}
func StartServer(i int, wg *sync.WaitGroup, rft *RaftType) {
	ch := make(chan int)
	server,valid := InitServer(i, "./config.json")
	if valid{
		//log.Println("inside valid",valid,server)
		server.Start(rft,ch)
		go start(ch)
	}
	wg.Done()
	return
}

func start(ch chan int){
	for i:=1;i<6;i++{

	stdin := bufio.NewReader(os.Stdin)
	//ch <- int(stdin.ReadString('\n'))
	_ ,_= stdin.ReadByte()
	ch <- i
	}
}
