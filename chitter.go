package main

import (
    "os"
    "net"
    "bufio"
    "strconv"
    "fmt"
    "strings"
)

type userDetails struct {
  id string
  conn net.Conn
}

type message struct {
  sender string
  recipient string
  body string
}

var userList map[string] net.Conn         // keep track of connected users

var idAssignmentChan = make(chan string)  // used to give IDs to all users
var addUserChan = make(chan userDetails)  // used to add user to userList
var broadcastMsgChan = make(chan message) // used to broadcast public messages
var privateMsgChan = make(chan message)   // used to send private messages

func handleConnection(conn net.Conn) {
    b := bufio.NewReader(conn)
    clientID := <-idAssignmentChan

    // create a struct to represent a new user and pass it along to userList
    newUser := userDetails{id: clientID, conn: conn}
    addUserChan<-newUser

    for {
        line, err := b.ReadBytes('\n')
        if err != nil {
            conn.Close()
            break
        }

        lineStr := string(line)

        // if message has no colon, broadcast it...
        if (!strings.Contains(lineStr, ":")) {
          // ... to everyone in userList
          broadcastMsg := message{sender: clientID, body: string(line)}
          broadcastMsgChan<-broadcastMsg

        } else {
          // otherwise, parse commands by first colon
          lineSplit := strings.SplitN(lineStr, ":", 2)

          // strip leading and trailing whitespace
          command := strings.Trim(lineSplit[0], " ")

          // all:
          if (command == "all") {
            body := strings.Trim(lineSplit[1], " ")
            broadcastMsg := message{sender: clientID, body: body}
            broadcastMsgChan<-broadcastMsg

          // whoami:
          } else if (command == "whoami") {
            conn.Write([]byte("chitter: " + clientID + "\n"))

          // private message (e.g. 0:)
          } else if _, err := strconv.Atoi(command); err == nil {
            body := strings.Trim(lineSplit[1], " ")
            privateMsg := message{sender: clientID, recipient: command, body: body}
            privateMsgChan<-privateMsg

          }
        }
    }
}

func userListManager() {
  userList = make(map[string] net.Conn)

  for {
    select {
    case newUser := <-addUserChan:
      userList[newUser.id] = newUser.conn
    case broadcastMsg := <-broadcastMsgChan:
      for _, conn := range userList {
        conn.Write([]byte(broadcastMsg.sender + ": " + broadcastMsg.body))
      }
    case privateMsg := <-privateMsgChan:
      if conn, ok := userList[privateMsg.recipient]; ok {
        conn.Write([]byte(privateMsg.sender + ": " + privateMsg.body))
      }
    }
  }
}

func idManager() {
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

    go idManager()
    go userListManager()

    fmt.Println("Listening on port", os.Args[1])
    for {
        conn, _ := server.Accept()
        go handleConnection(conn)
    }
}
