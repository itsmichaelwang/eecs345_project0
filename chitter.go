package main

import (
    "os"
    "net"
    "bufio"
    "strconv"
    "fmt"
    "strings"
)

type user struct {
    id string
    conn net.Conn
}

type message struct {
    sender string
    recipient string
    body string
}

var userList map[string] net.Conn         // tracks all connected users

var idAssignmentChan = make(chan string)  // used to assign IDs to all new users
var addUserChan = make(chan user)         // used to add users to userList

var publicMsgChan = make(chan message)    // used to broadcast public messages
var privateMsgChan = make(chan message)   // used to send private messages

/*
* A new instance of this is created for each user that joins the chat
* it registers the user to userList and watches for input
*/
func handleConnection(conn net.Conn) {
    b := bufio.NewReader(conn)
    clientID := <-idAssignmentChan

    newUser := user{id: clientID, conn: conn}
    addUserChan<-newUser

    for {
        line, err := b.ReadBytes('\n')
        if err != nil {
            conn.Close()
            break
        }
        lineStr := string(line)

        // NO COLON - default behavior is to broadcast a public message
        if (!strings.Contains(lineStr, ":")) {
            publicMsg := message{sender: clientID, body: lineStr}
            publicMsgChan<-publicMsg

        // COLON - interpret input as commands
        } else {
            lineSplitByFirstColon := strings.SplitN(lineStr, ":", 2)
            command := strings.Trim(lineSplitByFirstColon[0], " ")  // note the whitespace stripping

            if (command == "all") {
                body := strings.Trim(lineSplitByFirstColon[1], " ")
                publicMsg := message{sender: clientID, body: body}
                publicMsgChan<-publicMsg

            } else if (command == "whoami") {
                conn.Write([]byte("chitter: " + clientID + "\n"))

            } else if _, err := strconv.Atoi(command); err == nil {
                body := strings.Trim(lineSplitByFirstColon[1], " ")
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
        case publicMsg := <-publicMsgChan:
            for _, conn := range userList {
                conn.Write([]byte(publicMsg.sender + ": " + publicMsg.body))
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
