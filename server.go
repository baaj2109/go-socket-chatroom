package main


import (
    "fmt"
    "net"
    "strings"
    "time"
)
func main() {
    listener, err:= net.Listen("tcp", "localhost:8080")
    if err != nil { 
        fmt.Println(err)
        return
    }
    
    connMap := make(map[net.Conn]string)

    for {
        conn, err := listener.Accept()
        if err != nil {
            fmt.Println(err)
            continue    
        }
        connMap[conn] = conn.RemoteAddr().String()
        notifyAllNewUserLogin(connMap[conn], connMap)
        go HandleConn(conn, connMap)
    }
}   


/// handle every connect
func HandleConn(conn net.Conn, connMap map[net.Conn]string) {

    buf := make([]byte, 1024)
    /// close conn
    defer handlerConnClose(conn, connMap, 1)

    /// create a thread to monitor overtime
    keepAlive := make(chan bool)
    go func(conn net.Conn) {
        for {
            select {
            case <- keepAlive:
            case <- time.After(1 * time.Minute):
                /// this connect overtime
                handlerConnClose(conn, connMap, 2)
                return
            }
        }
    }(conn)

    /// print helper list
    help(conn)
    for {
        n, err:= conn.Read(buf)
        if err!= nil {
            return
        }
        keepAlive <- true
        msg := strings.Trim(string(buf[:n]), "\r\n")
        handleMsg(conn,connMap, msg)
    }
}

/// handle close connect
func handlerConnClose(conn net.Conn, connMap map[net.Conn]string, t int) {
    /// t = 1: initiative close connect
    /// t = 2: overtime close connect
    userExit(conn, connMap, t)
    _ = conn.Close()
}

/// notify new user login
func notifyAllNewUserLogin(name string, connMap map[net.Conn]string) {
    for curConn := range connMap {
        msg:= "[" +time.Now().Format("00:00:00") + "] - " + name + " login \n"
        _, err := curConn.Write([]byte(msg))
        if err != nil {
            fmt.Printf("failed to remind %s new user loging", connMap[curConn])
            continue
        }
    }
}

///notify all user someone logout
func notifyAllUserLogout(name string, connMap map[net.Conn]string, t int) {
    for curConn := range connMap {
        msg := "[" +time.Now().Format("00:00:00") + "] - " + name 
        if t == 1 {
            msg = msg + " logout \n"
        } else if t == 2 {
            msg = msg + " timeout \n"
        }
        _, err := curConn.Write([]byte(msg))
        if err != nil {
            fmt.Printf("failed to remind %s someone logout", connMap[curConn])
            continue
        }
    }
}

/// handle msg
func handleMsg(conn net.Conn, connMap map[net.Conn]string, msg string) {
    parseArr := strings.Split(msg, "|")
    if len(parseArr) > 1 && parseArr[0] == "func" {
        switch parseArr[1] {
        case "rename":
            {
                if len(parseArr) == 3 {
                    rename(conn, connMap, parseArr[2])
                } else {
                    syntaxError(conn)
                }
            }
        case "list":
            {
                if len(parseArr) == 2 {
                    list(conn, connMap)
                } else {
                syntaxError(conn)
                }
            }
        case "exit":
            {
                if len(parseArr) == 2 {
                    userExit(conn, connMap, 1)
                } else {
                    syntaxError(conn)
                }
            }
        }
    } else {
        broadcast(conn, connMap, msg)
    }
}

func broadcast(conn net.Conn, connMap map[net.Conn]string, msg string) {
    for curConn := range connMap {
        msg:= "[" +time.Now().Format("00:00:00") + "] - " + connMap[conn] + ": " + msg + "\n"
        _, err := curConn.Write([]byte(msg))
        if err != nil {
            fmt.Printf("failed to send msg to  %s", connMap[curConn])
            continue
        }
    }
}

/// logout
func userExit(conn net.Conn, connMap map[net.Conn]string, t int) {
    name, exist := connMap[conn]
    if exist {
        delete(connMap, conn)
        notifyAllUserLogout(name, connMap, t)
    }
}

func syntaxError(conn net.Conn) {
    _, err := conn.Write([]byte("syntaxError \n"))
    if err != nil {
        fmt.Printf("failed to send msg to %s, cause syntaxError", conn)
        return
    }
}


func rename(conn net.Conn, connMap map[net.Conn]string, newName string) {
    _, err := conn.Write([]byte("success \n"))
    if err != nil {
        fmt.Printf("failed to rename %s", conn)
        return
    }
    connMap[conn] = newName
}

func list(conn net.Conn, connMap map[net.Conn]string) {
    msg := "------------------\n"
    msg += "User List: \n"
    for curConn, name:= range connMap {
        
        if conn != curConn {
            msg += name + "\n"
        }
    }
    msg += "------------------\n"
    _, err := conn.Write([]byte(msg))
    if err != nil {
        fmt.Printf("failed to send user list to %s", conn)
        return
    }
}

func help(conn net.Conn) {
    msg := `
    --------------------------------------------------------------------------------------
    Introduction:
    This is a easy chat room!
    You can chat with other people by this client!
    As you input anything, other people in the room can see it, also, you can If you do nothing within 3 minutes, the client will go offline!
    Command List:
    <anything> func [help
    func| rename |<Your New Name> func|list func exit
    - chat with others
    - get some help change your name
    - to see who is online
    - go offline
    --------------------------------------------------------------------------------------
    `
    _, err := conn.Write([]byte(msg))
    if err != nil {
        fmt.Printf("failed to send msg to %s", conn)
        return
    }
}





























