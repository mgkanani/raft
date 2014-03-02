package raft

import (
	"os/exec"
	"sync"
	"testing"
	"time"
	"fmt"
	"strconv"
	"os"
)
//type Cmd os.exec.Cmd

var total_servers = 7
var cmd map[int]*exec.Cmd

func TestRaft(t *testing.T) {
	total_servers +=1
	wg := new(sync.WaitGroup)
	cmd = make(map[int]*exec.Cmd)
	path := os.Getenv("GOPATH")+"/bin/RaftMain"
	//fmt.Println("GOPATH=",path);
	for i := 1; i < total_servers; i++ {
		wg.Add(1)
		//go StartServer(i, wg)
		//str:= string(path)+"RaftMain -pid="+strconv.Itoa(i)
		cmd[i]=exec.Command(path,"-pid",strconv.Itoa(i));
		go start(cmd[i])
		//o,err:=cmd[i].Output()
		//fmt.Println("command is:",str,cmd[i],string(o),err)
	}
	go killProc(wg)
	wg.Wait()
	return

}
func start(cmd *exec.Cmd){
	err:=cmd.Run()
        fmt.Println("command is:",cmd,err)
}
/*
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
*/
func killProc(wg *sync.WaitGroup){
	time.After(5*time.Second)
	for i := 1; i < total_servers; i++ {
		time.After(5*time.Second)
		//_,ok:=cmd[i]
		fmt.Println(cmd[i].ProcessState)
		if !cmd[i].ProcessState.Exited(){
			time.After(5*time.Second)
			cmd[i].Process.Kill()
			wg.Done();
		}
	}
}
