package KeyValue

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"testing"
	"time"
)

var mutex1 = &sync.Mutex{}
var mtx_run = &sync.Mutex{}
var total_servers = 7
var cmd map[int]*exec.Cmd

var running string

type DevNull struct{}

func (DevNull) Write(p []byte) (int, error) {
	return len(p), nil
}

func TestRaft(t *testing.T) {
	//Read parametes from configuration file.
	file, e := ioutil.ReadFile(CONFIG)
	if e != nil {
		if debug {
			log.Printf("File error: %v\n", e)
		}
		os.Exit(1)
	}
	//decoding file's input and save data into variables.
	var obj jsonobject
	err := json.Unmarshal(file, &obj)
	if err != nil {
		if debug {
			log.Println("error:", err)
		}
	}
	Servers = make(map[int]string)
	for _, value := range obj.Object.Servers {
		Servers[value.MyPid] = value.Socket
		running = value.Socket
	}
	total_servers = len(Servers)

	log.SetOutput(new(DevNull))
	total_servers += 1
	wg := new(sync.WaitGroup)
	cmd = make(map[int]*exec.Cmd)
	path := os.Getenv("GOPATH") + "/bin/KeyValue"
	str := "KeyValue.log"
	outfile, err := os.Create(str)
	if err != nil {
		panic(err)
	}
	defer outfile.Close()
	for i := 1; i < total_servers; i++ {
		cmd[i] = exec.Command(path, strconv.Itoa(i))
		cmd[i].Stdout = outfile
		cmd[i].Stderr = outfile
		if debug {
			log.Println(cmd[i])
		}
	}
	for i := 1; i < total_servers; i++ {
		mtx_run.Lock()
		running = "http://" + Servers[i] + "/"
		mtx_run.Unlock()
		go start(cmd[i])
	}
	go start_client()
	go killProc(wg)
	wg.Add(1)
	wg.Wait()
	time.Sleep(8 * time.Second)
	for i := 1; i < total_servers; i++ {
		if debug {
			log.Println("process state:-", cmd[i].ProcessState)
		}
		mutex1.Lock()
		if cmd[i].Process != nil {
			if cmd[i].ProcessState == nil {
				if debug {
					log.Println("Kill process:-", cmd[i])
				}
			}
		}
		mutex1.Unlock()
	}
	return
}

func start(cmd *exec.Cmd) {
	err := cmd.Run()
	if debug {
		if !cmd.ProcessState.Success() {
			log.Println("error occured in running command", cmd.Path, cmd.Args)
		}
		log.Println("err is:", err)
		log.Println("command was:", cmd, "\tstate:-", cmd.ProcessState, "\t process:-", cmd.Process)
	}
}
func killProc(wg *sync.WaitGroup) {
	time.Sleep(8 * time.Second)
	//        for k := 0; k < 1; k++ {
	for i := 1; i < total_servers; i++ {
		if debug {
			log.Println("process state:-", cmd[i].ProcessState)
		}
		mutex1.Lock()
		if cmd[i].Process != nil { //process either exited successfully or in running status.
			if cmd[i].ProcessState == nil {
				if debug {
					log.Println("Kill process:-", cmd[i])
				}
				cmd[i].Process.Kill()
				//cmd[i].Process.Wait()
				time.Sleep(5 * time.Second)
				temp := &exec.Cmd{Path: cmd[i].Path, Args: cmd[i].Args, Stdout: cmd[i].Stdout, Stderr: cmd[i].Stderr}
				delete(cmd, i)
				cmd[i] = temp
				cmd[i].Start()
				mtx_run.Lock()
				running = "http://" + Servers[i] + "/"
				mtx_run.Unlock()
			}
		}
		mutex1.Unlock()
	}
	//        }
	wg.Done()
}

func Send(msg *Message, url string) {
	buf, _ := json.Marshal(msg)
	body := bytes.NewBuffer(buf)
	r, e := http.Post(url, "text/json", body)
	if e == nil {
		if debug {
			log.Println(r)
		}
		if r.Request.Method == "GET" {
			if debug {
				log.Println("Redirected to:-", r.Request.URL.String(), e, body)
			}
			Send(msg, r.Request.URL.String())
		} else {
			if debug {
				log.Println("Not redirected:-", e, r.Request.Method, r.Request.URL, body)
			}
			response, _ := ioutil.ReadAll(r.Body)
			if debug {
				log.Println(string(response), r.Header, "if err:-", e)
			}
		}
	} else {
		if debug {
			log.Println("error:-", e)
		}
	}
}

func start_client() {

	for i := 1; i < 250; i++ {
		go Set(i)
		go Get(i)
		go Update(i)
		go Dlt(i)
		time.Sleep(500 * time.Millisecond)
	}
}
func Set(i int) {
	mtx_run.Lock()
	temp_running := running
	mtx_run.Unlock()

	str_key := strconv.Itoa(i) + "abc" + strconv.Itoa(i)
	str_value := strconv.Itoa(i) + "testing" + strconv.Itoa(i)
	msg := &Message{Type: SET, Key: str_key, Value: str_value}
	buf, _ := json.Marshal(msg)
	body := bytes.NewBuffer(buf)
	if debug {
		log.Println(body)
	}
	Send(msg, temp_running)
}
func Get(i int) {
	mtx_run.Lock()
	temp_running := running
	mtx_run.Unlock()

	str_key := strconv.Itoa(i) + "abc" + strconv.Itoa(i)
	str_value := strconv.Itoa(i) + "testing" + strconv.Itoa(i)
	msg := &Message{Type: GET, Key: str_key, Value: str_value}
	buf, _ := json.Marshal(msg)
	body := bytes.NewBuffer(buf)
	if debug {
		log.Println(body)
	}
	Send(msg, temp_running)
}
func Update(i int) {
	mtx_run.Lock()
	temp_running := running
	mtx_run.Unlock()

	str_key := strconv.Itoa(i) + "abc" + strconv.Itoa(i)
	str_value := strconv.Itoa(i) + "testing" + strconv.Itoa(i)
	msg := &Message{Type: UPDATE, Key: str_key, Value: str_value}
	buf, _ := json.Marshal(msg)
	body := bytes.NewBuffer(buf)
	if debug {
		log.Println(body)
	}
	Send(msg, temp_running)
}
func Dlt(i int) {
	mtx_run.Lock()
	temp_running := running
	mtx_run.Unlock()

	str_key := strconv.Itoa(i) + "abc" + strconv.Itoa(i)
	str_value := strconv.Itoa(i) + "testing" + strconv.Itoa(i)
	msg := &Message{Type: DELETE, Key: str_key, Value: str_value}
	buf, _ := json.Marshal(msg)
	body := bytes.NewBuffer(buf)
	if debug {
		log.Println(body)
	}
	Send(msg, temp_running)
}
