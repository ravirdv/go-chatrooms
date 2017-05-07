package main

import (
        "log"
        "net/http"
        "github.com/gorilla/websocket"
)

// Define our message object
type Message struct {
        Username string `json:"username"`
        Message  string `json:"message"`
}


// list of connect users
var clients = make(map[*websocket.Conn]bool)
// channel for broadcasting messages
var broadcast = make(chan Message)

// Configure the upgrader
var upgrader = websocket.Upgrader{}

func main() {
        // Create a simple file server
        fs := http.FileServer(http.Dir("./public"))
        http.Handle("/", fs)
	// setup ws route
        http.HandleFunc("/ws", handleConnections)
	// start listening
	go handleMessages()
        // Start the server on localhost port 80
        log.Println("http server started on :80")
        err := http.ListenAndServe(":80", nil)
        if err != nil {
                log.Fatal("ListenAndServe: ", err)
        }
}


func handleConnections(w http.ResponseWriter, r *http.Request) {
	// We got a new conncetion
        log.Printf("We got a new connection %v", r)
        // Upgrade initial GET request to a websocket
        ws, err := upgrader.Upgrade(w, r, nil)
        if err != nil {
                log.Fatal(err)
        }
        // Register our new client
        clients[ws] = true
	for {
                var msg Message
                // Read in a new message as JSON and map it to a Message object
                err := ws.ReadJSON(&msg)
                if err != nil {
                        log.Printf("error: %v", err)
                        delete(clients, ws)
                        break
                }
                // Send the newly received message to the broadcast channel
                broadcast <- msg
        }
        // Make sure we close the connection when the function returns
        defer ws.Close()
}


func handleMessages() {
        for {
                // Grab the next message from the broadcast channel
                msg := <-broadcast
                // Send it out to every client that is currently connected
                for client := range clients {
                        err := client.WriteJSON(msg)
                        if err != nil {
                                log.Printf("error: %v", err)
                                client.Close()
                                delete(clients, client)
                        }
                }
        }
}
