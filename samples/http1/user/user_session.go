package user

import (
	"context"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 1024
)

type User struct {
	Email string
}

type UserSession struct {
	startTime    time.Time
	endTime      time.Time
	messageCount int64
	conn         *websocket.Conn
}

func NewSession() *UserSession {
	return &UserSession{startTime: time.Now(), endTime: time.Now()}
}

func (u *UserSession) Start(ctx context.Context, cancel context.CancelFunc, conn *websocket.Conn) error {
	// MAYBE: check deadline, if valid start the runloop
	u.conn = conn
	go u.keepAlive(ctx) // send ping/pong messages?
	u.runLoop(cancel)   // read/write message loop

	return nil
}

func (u *UserSession) keepAlive(ctx context.Context) {
	// check the expiry status for each device, and remove is they are expired
	c := time.Tick(5 * time.Second)
	for {
		select {
		case <-c:
			//u.sendKeepAlive
			u.conn.WriteMessage(websocket.TextMessage, []byte("keepalive"))
		case <-ctx.Done():
			break
		}
	}
}

func (u *UserSession) runLoop(cancel context.CancelFunc) error {
	u.conn.SetReadLimit(maxMessageSize)
	u.conn.SetReadDeadline(time.Now().Add(pongWait))
	u.conn.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		mt, message, err := u.conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			cancel() // Cancel the main context
			break
		}
		log.Printf("recv: %s", message)
		log.Println("Mt:", mt)
		err = u.conn.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			cancel() // Cancel the main context
			break
		}
	}

	return nil
}

func (u *UserSession) writeToConn(mt int, msg []byte) {
	u.conn.SetWriteDeadline(writeWait)
	u.conn.WriteMessage(mt, msg)
}
