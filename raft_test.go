package raft

import (
	"fmt"
	"io"
	"log"
	"net/rpc"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"testing"
	"time"
)

var mutex1 = &sync.Mutex{}
var total_servers = 7
var cmd map[int]*exec.Cmd
var dbg = false

type RPC_Msg struct {
	Term   int
	Leader int
}

type DevNull struct{}

func (DevNull) Write(p []byte) (int, error) {
	return len(p), nil
}

func TestRaft(t *testing.T) {
	log.SetOutput(new(DevNull))
	total_servers += 1
	wg := new(sync.WaitGroup)
	//temp := exec.Command("ls")
	//out,e:= temp.Output()
	//fmt.Println(string(out),e)
	cmd = make(map[int]*exec.Cmd)
	path := os.Getenv("GOPATH") + "/bin/RaftMain"
	str := "raft_servers.log"
	outfile, err := os.Create(str)
	if err != nil {
		panic(err)
	}
	defer outfile.Close()
	//fmt.Println("GOPATH=",path);
	for i := 1; i < total_servers; i++ {
		wg.Add(3)
		cmd[i] = exec.Command(path, "-pid", strconv.Itoa(i))
		/*
			str:="log_"+strconv.Itoa(i)
			outfile, err := os.Create(str)
			if err != nil {
			       	panic(err)
			}
			defer outfile.Close()
		*/
		cmd[i].Stdout = outfile
		cmd[i].Stderr = outfile
		if dbg {
			fmt.Println(cmd[i])
		}
	}
	for i := 1; i < total_servers; i++ {
		go start(cmd[i])
	}

	go checkingLeader()

	go killProc(wg)

	wg.Wait()

	for i := 1; i < total_servers; i++ {
		if dbg {
			fmt.Println("process state:-", cmd[i].ProcessState)
		}
		mutex1.Lock()
		if cmd[i].Process != nil {
			if cmd[i].ProcessState == nil {
				if dbg {
					fmt.Println("Kill process:-", cmd[i])
				}
				cmd[i].Process.Kill()
			}
		}
		mutex1.Unlock()
	}
	return

}

func checkingLeader() {
	var reply = make(map[int]*RPC_Msg)
	leaders := make(map[int]map[int]int)
	for {

		id := 21340
		for i := 1; i < total_servers; i++ {
			reply[i] = &RPC_Msg{}
			id += 1
			str := string("127.0.0.1:" + strconv.Itoa(id))
			client, err := rpc.Dial("tcp", str)
			if err != nil && (err != io.ErrUnexpectedEOF || err != io.EOF) {
				//fmt.Println("dialing:", err)
			} else {
				// Synchronous call
				err = client.Call("Test.GetStatus", &i, reply[i])
				if err != nil {
					if dbg {
						fmt.Println("GetSTatus error:", err)
					}
				} else {
					//fmt.Println(i,reply[i] ,str)
					if reply[i].Leader > 0 {
						_, ok := leaders[reply[i].Term]
						if !ok {
							leaders[reply[i].Term] = make(map[int]int)
						}
						leaders[reply[i].Term][reply[i].Leader] = reply[i].Leader
						if len(leaders[reply[i].Term]) > 1 {
							panic("more than One Leader")
						}
					}
				}
			}
		}
		//fmt.Println(leaders);
		time.Sleep(1500 * time.Millisecond)
	}
}

func start(cmd *exec.Cmd) {
	//err := cmd.Run()
	err := cmd.Run()
	if dbg {
		if !cmd.ProcessState.Success() {
			fmt.Println("error occured in running command", cmd.Path, cmd.Args)
		}
		fmt.Println("err is:", err)
		fmt.Println("command was:", cmd, "\tstate:-", cmd.ProcessState, "\t process:-", cmd.Process)
	}
}
func killProc(wg *sync.WaitGroup) {
	time.Sleep(5 * time.Second)
	for k := 0; k < 3; k++ {
		for i := 1; i < total_servers; i++ {
			//time.After(5 * time.Second)
			//_,ok:=cmd[i]
			if dbg {
				fmt.Println("process state:-", cmd[i].ProcessState)
			}
			mutex1.Lock()
			if cmd[i].Process != nil { //process either exited successfully or in running status.
				if cmd[i].ProcessState == nil {
					if dbg {
						fmt.Println("Kill process:-", cmd[i])
					}
					cmd[i].Process.Kill()
					//cmd[i].Process.Wait()
					time.Sleep(5 * time.Second)
					temp := &exec.Cmd{Path: cmd[i].Path, Args: cmd[i].Args, Stdout: cmd[i].Stdout, Stderr: cmd[i].Stderr}
					delete(cmd, i)
					cmd[i] = temp
					cmd[i].Start()
				}
			}
			mutex1.Unlock()
			wg.Done()
		}
	}
}
