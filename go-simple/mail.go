// websockets.go
package main

import (
    "fmt"
    "net/http"

    "github.com/gorilla/websocket"
    "github.com/friendsofgo/graphiql"

)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
}

func websocket_server(w http.ResponseWriter, r *http.Request){
    conn, _ := upgrader.Upgrade(w, r, nil) // error ignored for sake of simplicity

    for {
        // Read message from browser
        msgType, msg, err := conn.ReadMessage()
        if err != nil {
            return
        }

        // Print the message to the console
        fmt.Printf("%s sent: %s\n", conn.RemoteAddr(), string(msg))

        // Write message back to browser
        if err = conn.WriteMessage(msgType, msg); err != nil {
            return
        }
    }

}


func main() {
    http.HandleFunc("/echo",websocket_server)

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        http.ServeFile(w, r, "hello.html")
    })
	graphiqlHandler, err := graphiql.NewGraphiqlHandler("/graphql")
	if err != nil {
		panic(err)
	}

	http.Handle("/graphql", gqlHandler())
	http.Handle("/graphiql", graphiqlHandler)

    http.ListenAndServe(":8000", nil)
}
