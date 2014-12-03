package main

import (
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestCreateHttpServerConnection(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var upgrader = websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		}

		ws, err := upgrader.Upgrade(w, r, nil)
		assert.Nil(t, err)

		msg_type, data, err := ws.ReadMessage()
		assert.Equal(t, websocket.BinaryMessage, msg_type)
		assert.Equal(t, "REQUEST", string(data))

		assert.Nil(t, ws.WriteMessage(msg_type, []byte("RESPONSE")))
	}))
	defer ts.Close()

	server := httptest.NewServer(http.HandlerFunc(combiner(httpHandler, "ws"+ts.URL[4:])))
	defer server.Close()

	data := url.Values{}
	data.Set("msg", "REQUEST")
	response, err := http.PostForm(server.URL+"/qlm/", data)
	assert.Nil(t, err)
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t, "RESPONSE", string(body))
}
