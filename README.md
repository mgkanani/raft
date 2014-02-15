#raft
A `raft` cluster contains several servers which allows the system to tolerate  failure(for 5 servers, it will be 2). This template/pkg is implementing this algorithm. In this, one leader will be selected among the servers.


#Usage
## To install
```
go get github.com/mgkanani/raft
go install github.com/mgkanani/raft
```
## To run with default configuration:-
(go to cluster directory i.e. github.com/mgkanani/raft and run below to see msgs' summary).
```
go test
```

#To customise parameters
###Modify config.json

To add more servers with different port and pids.
Pids must be in strictly order means 1,2,3,4 but not 1,2,4,7.
The order 1,4,3,2 will work perfectly.

###Modifying raft_test.go


#Default configurations:-
```
Total Servers :- 5
ipaddr:127.0.0.1 
ports :-12345,12346,12347,12348,12349
Pids:-1,2,3,4,5
```

#Experiments:-
```
see log file.

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

#Mistakes(very silly but consumes much time) during coding which takes hours to find and resolve.

Making Request and Reply's data as private and using marshal and Unmarshal function which returns null data.


