package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"strconv"
)

func httpHandler(send, recv chan []byte, w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/qlm/" {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "404 page not found")
		return
	}
	if r.Method == "GET" {
		fmt.Fprintf(w, `<html><body><form action="." method="POST"><textarea name="msg" rows="10" cols="50"><?xml version="1.0" encoding="UTF-8"?><omi:omiEnvelope version="1.0" ttl="10"><omi:read><omi:msg></omi:msg></omi:read></omi:omiEnvelope></textarea><input type="submit" value="Submit"></form></body></html>`)
		return
	} else if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "405 Method Not Allowed")
		return
	}
	request := r.FormValue("msg")
	send <- []byte(request)
	response := <-recv
	fmt.Fprintf(w, string(response))
}
func wsServerConnector(address string) (chan []byte, chan []byte) {
	send := make(chan []byte)
	receive := make(chan []byte)
	go func() {
		for {
			select {
			case rawMsg := <-send:
				var h http.Header

				conn, _, err := websocket.DefaultDialer.Dial(address, h)
				if err == nil {
					if err := conn.WriteMessage(websocket.BinaryMessage, rawMsg); err != nil {
						receive <- []byte(err.Error())
					}
					_, content, err := conn.ReadMessage()
					if err == nil {
						receive <- content
					} else {
						receive <- []byte(err.Error())
					}
				} else {
					receive <- []byte(err.Error())
				}
			}
		}
	}()
	return send, receive
}

func combiner(f func(chan []byte, chan []byte, http.ResponseWriter, *http.Request), address string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		send, recv := wsServerConnector(address)
		f(send, recv, w, r)
	}
}

func main() {
	var port int
	flag.IntVar(&port, "port", 8080, "Listening port")

	var address string
	flag.StringVar(&address, "server", "ws://localhost:8000/qlm/", "QLM server address")

	flag.Parse()

	http.HandleFunc("/qlm/", combiner(httpHandler, address))
	http.ListenAndServe(":"+strconv.Itoa(port), nil)
}
