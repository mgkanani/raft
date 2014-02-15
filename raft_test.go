package raft

import (
/*	"flag"
	"bufio"
	"os"
	"fmt"
	Raft "github.com/mgkanani/raft"*/
	//  "io/ioutil"
        "sync"
	"testing"
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
	raftType:=&RaftType{Leader:0}

        wg := new(sync.WaitGroup)
        for i:=1;i<9;i++{
        wg.Add(1);
        go StartServer(i,wg,raftType)
        }
        wg.Wait();
        return



}
func StartServer(i int, wg *sync.WaitGroup,rft *RaftType){
        server := InitServer(i, "./config.json")
        server.Start(rft);
	wg.Done();
	return;
}
