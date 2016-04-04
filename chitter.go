package main

import (
    "os"
    "net"
    "bufio"
    "strconv"
    "fmt"
)

type userDetails struct {
  id string
  conn net.Conn
}

var userList map[string] net.Conn         // keep track of connected users

var idAssignmentChan = make(chan string)  // used to give IDs to all users
var addUserChan = make(chan userDetails)  // used to add user to userList

func HandleConnection(conn net.Conn) {
    b := bufio.NewReader(conn)
    clientID := <-idAssignmentChan

    // create a struct to represent a new user and pass it along to userList
    newUser := userDetails{id: clientID, conn: conn}
    addUserChan <- newUser

    for {
        line, err := b.ReadBytes('\n')
        if err != nil {
            conn.Close()
            break
        }

        // every time a message is sent, broadcast it to everyone in userList
        for _, v := range userList {
          v.Write([]byte(clientID + ": " +string(line)))
        }
    }
}

func userListManager() {
  userList = make(map[string] net.Conn)

  for {
    select {
    case newUser := <-addUserChan:
      userList[newUser.id] = newUser.conn
      fmt.Println(userList)
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
    go userListManager()

    fmt.Println("Listening on port", os.Args[1])
    for {
        conn, _ := server.Accept()
        go HandleConnection(conn)
    }
}
