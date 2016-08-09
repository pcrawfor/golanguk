package user

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/context"

	"github.com/gorilla/websocket"
	"github.com/pcrawfor/golanguk/samples/app/lookup"
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

func (u *UserSession) Start(ctx context.Context, conn *websocket.Conn) error {
	subCtx, cancel := context.WithCancel(ctx)
	// MAYBE: check deadline, if valid start the runloop
	u.conn = conn
	go u.keepAlive(subCtx)    // send ping/pong messages?
	u.runLoop(subCtx, cancel) // read/write message loop

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

func (u *UserSession) runLoop(ctx context.Context, cancel context.CancelFunc) error {
	defer cancel()

	//u.conn.SetReadLimit(maxMessageSize)
	//u.conn.SetReadDeadline(time.Now().Add(pongWait))
	//u.conn.SetPongHandler(func(string) error { u.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		mt, message, err := u.conn.ReadMessage()
		if err != nil {
			log.Println("read error:", err)
			break
		}
		log.Printf("recv: %s", message)
		log.Println("Mt:", mt)

		// interpret the message and pass it off to process
		resp, herr := u.handleMessage(ctx, message)
		if herr != nil {
			log.Println("Error parsing message", err)
			resp = []byte("Whoops couldn't handle that message!")
		}

		if len(resp) < 1 {
			resp = []byte("Looks like we couldn't find that one!")
		}

		fmt.Println("Resp:", string(resp))

		err = u.writeToConn(mt, resp) //u.conn.WriteMessage(mt, message)
		if err != nil {
			log.Println("write error:", err)
			break
		}
	}

	return nil
}

func (u *UserSession) writeToConn(mt int, msg []byte) error {
	//u.conn.SetWriteDeadline(time.Now().Add(writeWait))
	return u.conn.WriteMessage(mt, msg)
}

func (u *UserSession) handleMessage(ctx context.Context, message []byte) ([]byte, error) {
	str := string(message)
	reCmd := regexp.MustCompile(`/\w+\s`)
	reQry := regexp.MustCompile(`\s\w.*`)
	if strings.HasPrefix(str, "/") {
		cmd := reCmd.FindString(str)
		cmd = strings.Replace(cmd, " ", "", -1)
		qry := reQry.FindString(str)

		log.Println("CMD:", cmd)
		log.Println("QRY:", qry)

		switch cmd {
		case "/ask":
			log.Println("process ask command")
			return u.handleQuestion(ctx, qry)
		case "/gif":
			return u.handleGif(ctx, qry)
		}
	}

	return message, nil // echo it
}

func (u *UserSession) handleQuestion(ctx context.Context, qry string) ([]byte, error) {
	log.Println("Handle Question:", qry)
	r, e := lookup.DuckduckQuery(ctx, qry)
	fmt.Println("R:", r)
	return []byte(r), e
}

func (u *UserSession) handleGif(ctx context.Context, qry string) ([]byte, error) {
	g := lookup.NewGiphy("dc6zaTOxFJmzC")

	// split the query into terms
	terms := strings.Split(qry, " ")
	url, err := g.GifForTerms(ctx, terms)
	if err != nil {
		return []byte{}, err
	}

	return []byte(url), nil
}
