package gongular

import (
	"fmt"
	"net"
	"net/http"
	"testing"

	"io"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	websocket_go "golang.org/x/net/websocket"
)

type wsTest struct {
	Param struct {
		UserID int
	}
	Query struct {
		Track    bool
		Username string
	}
}

func (w *wsTest) Before(c *Context) (http.Header, error) {
	return nil, nil
}

func (w *wsTest) Handle(conn *websocket.Conn) {
	_, msg, err := conn.ReadMessage()
	if err != nil {
		conn.Close()
	}

	toSend := fmt.Sprintf("%s:%d:%s:%t", msg, w.Param.UserID, w.Query.Username, w.Query.Track)
	conn.WriteMessage(websocket.TextMessage, []byte(toSend))
}

func TestWS_Simple(t *testing.T) {
	e := newEngineTest()
	e.GetWSRouter().Handle("/ws1/:UserID", &wsTest{})

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	go func() {
		http.Serve(listener, e.GetHandler())
	}()

	url := fmt.Sprintf("ws://%s/ws1/5?Track=true&Username=musti", listener.Addr().String())
	origin := fmt.Sprintf("http://%s/", listener.Addr().String())
	ws, err := websocket_go.Dial(url, "", origin)
	require.NoError(t, err)

	_, err = io.WriteString(ws, "selam")
	require.NoError(t, err)

	var buf = make([]byte, 1024)
	n, err := ws.Read(buf)
	require.NoError(t, err)

	result := string(buf[:n])
	listener.Close()

	assert.Equal(t, "selam:5:musti:true", result)
}
