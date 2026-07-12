// websockets.go
package main

import (
	"fmt"
	"net/http"

	"github.com/friendsofgo/graphiql"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func websocketServer(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	for {
		// Echo messages back to clients until they disconnect.
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}

		fmt.Printf("%s sent: %s\n", conn.RemoteAddr(), string(msg))

		if err = conn.WriteMessage(msgType, msg); err != nil {
			return
		}
	}
}

func routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/echo", websocketServer)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "hello.html")
	})
	graphiqlHandler, err := graphiql.NewGraphiqlHandler("/graphql")
	if err != nil {
		panic(err)
	}

	mux.Handle("/graphql", gqlHandler())
	mux.Handle("/graphiql", graphiqlHandler)
	return mux
}

func main() {
	if err := http.ListenAndServe(":8000", routes()); err != nil {
		panic(err)
	}
}
