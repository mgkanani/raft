#raft
A `raft` cluster contains several servers which allows the system to tolerate  failure(for 5 servers, it will be 2). This template/pkg is implementing this algorithm. In this, one leader will be selected among the servers.


#Usage
## To install
```
go get github.com/mgkanani/cluster
go get github.com/mgkanani/raft
go install github.com/mgkanani/raft
```

##Documentation reference:-
[![GoDoc](https://godoc.org/github.com/mgkanani/raft?status.png)](https://godoc.org/github.com/mgkanani/raft)

## To run with default configuration:-

Till now partial test file has been developed. It is not fully automated for testing so it requires user's involvement.
(go to cluster directory i.e. github.com/mgkanani/raft and run below after some time use cntr+c to stop the execution and observe the log).
```
go test 2>log_file_name
```

(to see which server had been declared as a Leader use)
```
cat log_file_name|grep Declared
```


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

#Mistakes(very silly but consumes much time) during coding which takes hours to find and resolve.

Making Request and Reply's data as private and using marshal and Unmarshal function which returns null data.


