package main

import (
	"flag"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/MDGSF/utils/log"
	"github.com/antonholmquist/jason"
	"github.com/gorilla/websocket"
	"github.com/vmihailenco/msgpack"
)

// TAuth device --> receiver server
type TAuth struct {
	Type    string `msgpack:"type"`
	Device  string `msgpack:"device"`
	Token   string `msgpack:"token"`
	ICCID   string `msgpack:"iccid"`
	Version string `msgpack:"version"`
	Time    int64  `msgpack:"time"`
}

// TAuthRet receiver server --> device
type TAuthRet struct {
	Type   string `msgpack:"type"`
	Accept string `msgpack:"accept,omitempty"`
	Reason string `msgpack:"reason,omitempty"`
}

// TMsg device send TMsg to receiver.
type TMsg struct {
	Type   string        `msgpack:"type"` //use string "status, event, file"
	ID     string        `msgpack:"id"`   //uuid
	Source string        `msgpack:"source"`
	Data   []interface{} `msgpack:"data"`
}

// Ack when receiver server recv an TMsg, it will reposne an Ack to device.
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
	http.HandleFunc("/", Index)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func Index(w http.ResponseWriter, r *http.Request) {

	// upgrade http request to websocket.
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("upgrade:", err)
		return
	}
	defer func() {
		conn.Close()
	}()

	localAddr := conn.LocalAddr().String()
	remoteAddr := conn.RemoteAddr().String()

	log.Info("new conn comming, local = %v, remote = %v", localAddr, remoteAddr)
	defer func() {
		log.Info("connection closed, local = %v, remote = %v", localAddr, remoteAddr)
	}()

	// try to recv device client auth message.
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Error("read: err = %v", err)
			return
		}

		m := &TAuth{}

		if err := msgpack.Unmarshal(message, m); err != nil {
			log.Error("not auth message")
			continue
		}

		if m.Type != "auth" {
			log.Error("Invalid auth msg = %v", m)
			continue
		}

		log.Info("recv auth msg = %v", m)

		// send device id and token to tokenServer to verify is valid or not.
		rsp, err := http.PostForm("https://setup.minieye.cc/services/report/agent/token/verify",
			url.Values{"device": {m.Device}, "token": {m.Token}})
		if err != nil {
			log.Error("send verify message to token server failed. err = %v", err)
			continue
		}

		body, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			log.Error("read token server response body failed, err = %v", err)
			continue
		}
		rsp.Body.Close()

		jsonBody, err := jason.NewObjectFromBytes(body)
		if err != nil {
			log.Error("parse json from token server failed, err = %v", err)
			continue
		}

		result, err := jsonBody.GetBoolean("result")
		if err != nil || !result {
			errorMsg, err := jsonBody.GetString("error")
			if err != nil {
				SendAuthFailedToDevice(conn, "unknown error")
			} else {
				SendAuthFailedToDevice(conn, errorMsg)
			}

			continue
		}

		log.Info("device = %v auth success", m)
		SendAuthSuccesToDevice(conn)
		break
	}

	// after auth success, try to read event message, status message, file message and dms message from device.
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Error("read: err = %v", err)
			break
		}

		m := &TMsg{}
		msgpack.Unmarshal(message, m)

		if m.Type != "file" {
			log.Info("recv %s msg, msglen = %v, %v\n", m.Type, len(message), m)
		} else {
			log.Info("recv %s msg, msglen = %v\n", m.Type, len(message))
		}

		ack := &Ack{}
		ack.Type = "ack"
		ack.ID = m.ID
		data, _ := msgpack.Marshal(ack)
		if err := conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
			return
		}
	}
}

func SendAuthSuccesToDevice(conn *websocket.Conn) {
	ret := &TAuthRet{Type: "auth_ok", Accept: "all"}
	data, _ := msgpack.Marshal(ret)
	if err := conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
		return
	}
}

func SendAuthFailedToDevice(conn *websocket.Conn, reason string) {
	ret := &TAuthRet{Type: "auth_error", Reason: reason}
	data, _ := msgpack.Marshal(ret)
	if err := conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
		return
	}
}
