package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
        SET    = 1
        GET    = 2
        UPDATE = 3
        DELETE = 4
)


type Message struct {
	Type int //SET,GET,UPDATE,DELETE
	Key string
	Value  string
}

func main() {
	msg := &Message{Type:GET, Key:"abc",Value:"testing"}
	buf, _ := json.Marshal(msg)
	body := bytes.NewBuffer(buf)
	fmt.Println(body)
	//Send(body,"http://127.0.0.1:8081/")
	Send(msg,"http://127.0.0.1:45003/")
}

func Send(msg *Message,url string){
	buf, _ := json.Marshal(msg)
	body := bytes.NewBuffer(buf)
        r, e := http.Post(url, "text/json", body)
	if e == nil{
		fmt.Println(r)
	        if r.Request.Method == "GET"{
        	        fmt.Println("Redirected to:-",r.Request.URL.String(),e,body)
			Send(msg,r.Request.URL.String())
	        }else{
	        	fmt.Println("Not redirected:-",e,r.Request.Method,r.Request.URL,body)
        	        response, _ := ioutil.ReadAll(r.Body)
                	fmt.Println(string(response),r.Header,"if err:-",e)
		}
	}else{
		fmt.Println("error:-",e)
	}
}
