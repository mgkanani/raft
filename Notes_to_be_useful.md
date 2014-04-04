-> In implementation of http's POST request redirection, They are handling only for status code 302 and 303 but for GET 
request they are handling for 30x status code.

->zmq can hanlde message of any size.

->encoding/binay is used for int64 to []byte conversion.
Varint -> []byte to int64 conversion
PutVarint -> reverse of above.

->to maunally perform the test of KeyValue servers:-
    
    -start servers using 
            KeyValue pid
    -start client using 
            for i in {1..200}; do ./client; echo $i ;done



->os.exec package is used for executing commands.
There are two Variable in structure Cmd in os/exec "Process" and "ProcessState" which can be used to know process state or exit status etc..
Below are the cases which can be useful during using this.

"Status of Process"		ProcessState-variable's content			Process-Variable's content

error				nil						nil

running				nil						non-nil

exited/completed		non-nil						non-nil
