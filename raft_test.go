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
	//"log"
)

var total_servers = 7

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
//	raftType := &RaftType{Leader: 0}
	total_servers +=1
	wg := new(sync.WaitGroup)
	for i := 1; i < total_servers; i++ {
		wg.Add(1)
		go StartServer(i, wg)
	}
	wg.Wait()
	return

}
func StartServer(i int, wg *sync.WaitGroup) {
	ch := make(chan int)
	valid := InitServer(i, "./config.json",true , ch)
	if valid{
		//log.Println("inside valid",valid,server)
		go start(ch)
	}
	//wg.Done()
	return
}

func start(ch chan int){
	for i:=1;i<total_servers;i++{

	stdin := bufio.NewReader(os.Stdin)
	//ch <- int(stdin.ReadString('\n'))
	_ ,_= stdin.ReadByte()
	ch <- i
	}
}
