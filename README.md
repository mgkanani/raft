#cluster
`cluster` is template for creating serverfarm. This is allowing to communicate in between them. One can either broadcast the message or send to individual server lying within `cluster`.


#Usage
## To install
```
go get github.com/mgkanani/cluster
go install github.com/mgkanani/cluster
```
## To run with default configuration:-
(go to cluster directory i.e. github.com/mgkanani/cluster and run below to see msgs' summary).
```
go test
```

#To customise parameters
###Modify config.json

To add more servers with different port and pids.
Pids must be in strictly order means 1,2,3,4 but not 1,2,4,7.
The order 1,4,3,2 will work perfectly.

###Modifying cluster_test.go

for Testing update values of variables total_servers count , and total_msgs.

for testing the msgs tranfer of size > 60000 bytes uncomment lines 48,57 and comment line 49.

#Default configurations:-
```
Total Servers :- 10
ipaddr:127.0.0.1 
ports :-12345,12346,12347,12348 ... 12354
Pids:-1,2,3,4 ... 10
total_servers=10(linenum 18 in file cluster_test.go),
total_msgs=300(linenum 19 in file cluster_test.go)
```

#Experiments:-
```
totalnumber of servers:-10
total msgs each Server broadcasts :-300
msg_size:-65K bytes

```
###output
```
total msgs received at PID - 1 :- 2700 	 sent:- 300
total msgs received at PID - 2 :- 2700 	 sent:- 300
total msgs received at PID - 3 :- 2700 	 sent:- 300
total msgs received at PID - 4 :- 2700 	 sent:- 300
total msgs received at PID - 5 :- 2700 	 sent:- 300
total msgs received at PID - 6 :- 2700 	 sent:- 300
total msgs received at PID - 7 :- 2700 	 sent:- 300
total msgs received at PID - 8 :- 2700 	 sent:- 300
total msgs received at PID - 9 :- 2700 	 sent:- 300
total msgs received at PID - 10 :- 2700 	 sent:- 300
PASS
ok  	github.com/mgkanani/cluster	333.525s

```



#References for basics of ZeroMQ and implemention

use of offline documentation of go-language that is using below commands :-
to find path of offline documentation:- ```godoc dfdsgdg ``` this package will not available and it will tell the path of godoc is looking for.
to find  available packages:- ```ls  /usr/lib/go/src/pkg``` , now observe packages and find approapriate pkg that is math/rand.
now use:-```godoc math/rand ```

[https://ramcloud.stanford.edu/wiki/download/attachments/11370504/raft.pdf](https://ramcloud.stanford.edu/wiki/download/attachments/11370504/raft.pdf) -- for reading raft paper as.
