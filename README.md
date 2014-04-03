#raft
A `raft` cluster contains several servers which allows the system to tolerate  failure(for 5 servers, it will be 2). This template/pkg is implementing this algorithm. In this, one leader will be selected among the servers.


#Implementation Details
-Key-Value + Raft + leveldb
-On startup Key-Value Map will be constructed from Logs saved in leveldb

#Extension
-if request comes to follower,redirecting it to leader.
-use of zmq/http(redirect will be simpler) instead of relying on simple tcp client-server.
-automate test for Key-Value.

#Dependencies:-
LevelDB:- github.com/syndtr/goleveldb/leveldb
ZMQ:- github.com/pebbe/zmq4

#Usage
## To install
```
go get github.com/mgkanani/cluster
go get github.com/mgkanani/raft
go install github.com/mgkanani/raft
go install github.com/mgkanani/raft/KeyValue
```


## TO Test
```
go get github.com/mgkanani/raft/RaftMain
go install github.com/mgkanani/raft/RaftMain
go test github.com/mgkanani/raft
```


## Tests performed:-
-Tests are performed on 3-different machines(From 7 servers 2,2,3 servers on respective machines) for testing the working of Raft.

-```go test ``` performs testing of raft by killing each process one by one after every 15-sec. This will be performed 3 times.


-2 clients sending 200 requests sending simultaneously, 7-servers were running on single host. Leader was killed manually[Cntr+c] when it reached to only 3 servers, there was no leader, then one by one server started. Then Leveldb data checked and found identical.



## To modify certain parameters  OR  Detailed Testing:-
```
to see all logs and what is happening during execution,
      modify RaftMain/main.go file,line-37 update to true instead of false
      modify raft_test.go,line-19, do same as above.
      

What is tested during testing or how testing is performed on servers:-
  -All servers are started
  -at every 5-sec 1-server will go down
  -all servers exactly 3-times goes down during testing.
  -at every 1.5sec who is Leader is checked and checked whether there are  more than one Leader.
      If there is more than one Leader is found , then test will return panic.
```


##Documentation reference:-
[![GoDoc](https://godoc.org/github.com/mgkanani/raft?status.png)](https://godoc.org/github.com/mgkanani/raft)


###Modify config.json

To add more servers with different port and pids.
Pids must be in strictly order means 1,2,3,4 but not 1,2,4,7.
The order 1,4,3,2 will work perfectly.

###Modifying raft_test.go


#Default configurations:-
```
Total Servers :- 8
ipaddr:127.0.0.1 
ports :-12345,12346,12347,12348,12349,... ,12352
Pids:-1,2,3,4,5
```

#Experiments:-
```
see log files.

log -> for 5 servers with 15sec sleep for leader after declared as a Leader.

log_... -> name itself suggests the configuration parameters.

Experiments are done with 8,9,5,and 7 servers with random and fix sleeping time after becoming a Leader. With these experiments it works correctly.


```



#References for basics of ZeroMQ and implemention

use of offline documentation of go-language that is using below commands :-
to find path of offline documentation:- ```godoc dfdsgdg ``` this package will not available and it will tell the path of godoc is looking for.
to find  available packages:- ```ls  /usr/lib/go/src/pkg``` , now observe packages and find approapriate pkg that is math/rand.
now use:-```godoc math/rand ```

[https://ramcloud.stanford.edu/wiki/download/attachments/11370504/raft.pdf]( -- for reading raft paper as).

[http://stackoverflow.com/questions/8270816/converting-go-struct-to-json](http://stackoverflow.com/questions/8270816/converting-go-struct-to-json)

[https://groups.google.com/forum/#!msg/golang-dev/oZdV_ISjobo/N-vfSnrcqhcJ]

[http://golang.org/pkg/time/#Timer.Stop]

[http://stackoverflow.com/questions/13812121/how-to-clear-a-map-in-go]

[http://stackoverflow.com/questions/1841443/iterating-over-all-the-keys-of-a-golang-map]

[http://stackoverflow.com/questions/11820842/how-to-configure-golang-so-it-can-access-environment-variables-in-osx]

[http://stackoverflow.com/questions/19965795/go-golang-write-log-to-file]

[http://stackoverflow.com/questions/18986943/in-golang-how-can-i-write-the-stdout-of-an-exec-cmd-to-a-file]

[http://www.sunzhongkui.me/rpc-communication-in-go-language/]

[http://play.golang.org/p/5LIA41Iqfp] -> to discard all log data ( this is done only for raft_test.go file.)

[http://stackoverflow.com/questions/21532113/golang-converting-string-to-int64]

#Mistakes(very silly but consumes much time) during coding which takes hours to find and resolve.

Making Request and Reply's data as private and using marshal and Unmarshal function which returns null data.


