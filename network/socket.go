package network

import (
	. "chat_server_golang/types"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  SocketBufferSize,
	WriteBufferSize: MessageBufferSize,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// Room 채팅방의 대한 값을 가지고 있음
type Room struct {
	Forward chan *message // 수신되는 메세지를 보관하는 값
	// 들어오는 메세지를 다른 클라이 언트 들에게 전송한다.

	Join  chan *client // 소켓이 연결되는 경우에 작동
	Leave chan *client // 소켓이 끊어지는 경우에 작동

	Clients map[*client]bool // 현재 방에 있는 client 정보를 저장
}

type message struct {
	Name    string
	Message string
	Time    int64
}

type client struct {
	Send   chan *message
	Room   *Room
	Name   string
	Socket *websocket.Conn
}

// NewRoom Room 이라는 객체값을 만들어 줄 수 있는 함수 작성
func NewRoom() *Room {
	return &Room{
		Forward: make(chan *message),
		Join:    make(chan *client),
		Leave:   make(chan *client),
		Clients: make(map[*client]bool)}
}

// Read 에서 계속해서 무한루프가 돌고 있기 때문에 SocketServe 에서 커넥션이 끊기지 않음
func (c *client) Read() {
	// 클라이언트가 들어오는 메세지를 읽는 함수
	defer c.Socket.Close()
	for {
		var msg *message
		err := c.Socket.ReadJSON(&msg) // unmarshal
		if err != nil {
			if !websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				break
			} else {
				panic(err)
			}
		} else {
			log.Println("READ : ", msg, "client", c.Name)
			log.Println()
			msg.Time = time.Now().Unix()
			msg.Name = c.Name

			c.Room.Forward <- msg
		}
	}
}

func (c *client) Write() {
	defer c.Socket.Close()
	// 클라이언트가 들어오는 메세지를 전송하는 함수

	for msg := range c.Send {
		log.Println("WRITE : ", msg, "client", c.Name)
		log.Println()
		err := c.Socket.WriteJSON(msg)
		if err != nil {
			panic(err)
		}
	}
}

// 외부에서 호출할 함수
func (r *Room) RunInit() {
	// Room 에 있는 모든 채널값들을 받는 역할
	for {
		select {
		case client := <-r.Join:
			r.Clients[client] = true
		case client := <-r.Leave:
			r.Clients[client] = false
			close(client.Send) // 채널을 닫아주는 역할
			delete(r.Clients, client)
		case msg := <-r.Forward: // 모든 클라이언트에 메세지를 전파해야함
			for client := range r.Clients {
				client.Send <- msg
			}
		}
	}
}

// gin 의 기본적인 인터페이스 구조라고 생각하면 됨.
func (r *Room) SocketServe(c *gin.Context) {
	socket, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		panic(err)
	}

	userCookie, err := c.Request.Cookie("auth")
	if err != nil {
		panic(err)
	}

	// user 의 대한 네이밍을 가져오고 Client 객체를 만들 수 있다.
	client := &client{
		Socket: socket,
		Send:   make(chan *message, MessageBufferSize),
		Room:   r,
		Name:   userCookie.Value,
	}

	r.Join <- client

	// 밑에 있는 로그까지 실행이 되고 defer 함수가 진행된다.
	defer func() { r.Leave <- client }()

	go client.Write() // 와 이건 뭐지

	client.Read()
}
