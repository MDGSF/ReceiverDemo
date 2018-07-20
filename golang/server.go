package main

import (
	"flag"
	"net/http"
	"sync"

	"github.com/MDGSF/utils/log"
	"github.com/gorilla/websocket"
	"github.com/vmihailenco/msgpack"
)

type TAuth struct {
	Type    string `msgpack:"type"`
	Device  string `msgpack:"device"`
	Token   string `msgpack:"token"`
	ICCID   string `msgpack:"iccid"`
	Version string `msgpack:"version"`
	Time    int64  `msgpack:"time"`
}

type TAuthRet struct {
	Type   string `msgpack:"type"`
	Accept string `msgpack:"accept"`
	Reason string `msgpack:"reason"`
}

type TMsg struct {
	Type   string        `msgpack:"type"` //use string "dsm"
	ID     string        `msgpack:"id"`   //uuid
	Source string        `msgpack:"source"`
	Data   []interface{} `msgpack:"data"`
}

type Ack struct {
	Type string `msgpack:"type"` // "ack"
	ID   string `msgpack:"id"`
	Time int64  `msgpack:"Time"`
}

var addr = flag.String("addr", "localhost:12306", "http service address")

var upgrader = websocket.Upgrader{} // use default options

func main() {
	flag.Parse()

	log.Info("MockReceiver start listen at %v", *addr)
	http.HandleFunc("/", echo)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func echo(w http.ResponseWriter, r *http.Request) {
	defer func() {
		log.Info("connection closed.")
	}()

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("upgrade:", err)
		return
	}

	log.Info("new conn comming")

	for {
		mt, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Error("read: err = %v", err)
			break
		}

		m := &TMsg{}
		msgpack.Unmarshal(message, m)

		if m.Type != "file" {
			log.Info("%v: %s, len = %v, %v\n", mt, m.Type, len(message), m)
		} else {
			log.Info("%v: %s, len = %v\n", mt, m.Type, len(message))
		}

		ack := &Ack{}
		ack.Type = "ack"
		ack.ID = m.ID
		data, _ := msgpack.Marshal(ack)
		if err := c.conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
			return
		}
	}

	client := NewClientConn(conn)
	client.Start()
	<-client.Closed
}

type TClientConn struct {
	conn   *websocket.Conn
	Closed chan bool
	once   sync.Once
}

func NewClientConn(conn *websocket.Conn) *TClientConn {
	c := &TClientConn{}
	c.conn = conn
	c.Closed = make(chan bool)
	return c
}

func (c *TClientConn) Close() {
	c.conn.Close()
	close(c.Closed)
}

func (c *TClientConn) Start() {
	go c.readFromClient()
}

func (c *TClientConn) readFromClient() {
	defer func() {
		c.once.Do(c.Close)
	}()

}
