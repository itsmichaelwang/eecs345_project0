package main

import (
    "os"
    "net"
    "bufio"
    "strconv"
    "fmt"
)

var idAssignmentChan = make(chan string)  // assigns IDs to all new users
var userList map[string] net.Conn         // maintains a list of connected users

func HandleConnection(conn net.Conn) {
    b := bufio.NewReader(conn)
    client_id := <-idAssignmentChan       // user id -> connection address

    userList[client_id] = conn

    for {
        line, err := b.ReadBytes('\n')
        if err != nil {
            conn.Close()
            break
        }

        // strings without colon should be broadcast by default
        // if !strings.Contains(line, ":") {
        //
        // }

        for _, v := range userList {
          v.Write([]byte(client_id + ": " +string(line)))
        }
    }
}

func IdManager() {
    var i uint64
    for i = 0;  ; i++ {
        idAssignmentChan <- strconv.FormatUint(i, 10)
    }
}

func main() {
    if len(os.Args) < 2{
        fmt.Fprintf(os.Stderr, "Usage: chitter <port-number>\n")
        os.Exit(1)
        return
    }
    port := os.Args[1]
    server, err := net.Listen("tcp", ":"+ port )
    if err != nil {
        fmt.Fprintln(os.Stderr, "Can't connect to port")
        os.Exit(1)
    }
    go IdManager()
    userList = make(map[string] net.Conn)

    fmt.Println("Listening on port", os.Args[1])
    for {
        conn, _ := server.Accept()
        go HandleConnection(conn)
    }
}
