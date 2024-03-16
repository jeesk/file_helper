package main

import (
	"file-helper/db"
	"fmt"
	"github.com/goccy/go-json"
	"github.com/lxzan/gws"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	PingInterval = 5 * time.Second
	PingWait     = 10 * time.Second
)

type FpHanlder struct {
	handler *Handler
}

const loginUUID = "uuid"

func (f FpHanlder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/fp/ws" || r.URL.Path == "/ws" {
		f.webSocketConn(w, r)
	} else if r.URL.Path == "/fp/all" || r.URL.Path == "/all" {
		// 返回所有的消息
		macIdStr := r.URL.Query().Get("minId")
		var minId = 0
		if macIdStr != "" {
			id, err := strconv.Atoi(macIdStr)
			if err != nil {
				w.Write([]byte("参数错误"))
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			minId = id
		}
		msg, err := db.MsgDB.QueryMsg(minId)
		if err != nil {
			w.Write([]byte("查询失败" + err.Error()))
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		marshal, _ := json.Marshal(msg)
		w.Write(marshal)
	}
}

func (f *FpHanlder) webSocketConn(writer http.ResponseWriter, request *http.Request) {
	upgrader := gws.NewUpgrader(f.handler, &gws.ServerOption{
		ParallelEnabled:   true,                                 // Parallel message processing
		Recovery:          gws.Recovery,                         // Exception recovery
		PermessageDeflate: gws.PermessageDeflate{Enabled: true}, // Enable compression
		Authorize: func(r *http.Request, session gws.SessionStorage) bool {
			var name = r.URL.Query().Get(loginUUID)
			if name == "" {
				return false
			}
			session.Store(loginUUID, name)
			session.Store("key", r.Header.Get("Sec-WebSocket-Key"))
			return true
		},
	})
	socket, err := upgrader.Upgrade(writer, request)
	if err != nil {
		return
	}
	go func() {
		socket.ReadLoop()
	}()
}

type Handler struct {
	sessions *gws.ConcurrentMap[string, *gws.Conn]
}

func (c *Handler) getName(socket *gws.Conn) string {
	name, _ := socket.Session().Load(loginUUID)
	return name.(string)
}

func (c *Handler) getKey(socket *gws.Conn) string {
	name, _ := socket.Session().Load("key")
	return name.(string)
}

func (c *Handler) OnOpen(socket *gws.Conn) {
	name := c.getName(socket)
	if conn, ok := c.sessions.Load(name); ok {
		conn.WriteClose(1000, []byte("connection replaced"))
	}
	socket.SetDeadline(time.Now().Add(PingInterval + PingWait))
	c.sessions.Store(name, socket)
	log.Printf("%s connected\n", name)
}

func (c *Handler) OnClose(socket *gws.Conn, err error) {
	name := c.getName(socket)
	key := c.getKey(socket)
	if mSocket, ok := c.sessions.Load(name); ok {
		if mKey := c.getKey(mSocket); mKey == key {
			c.sessions.Delete(name)
		}
	}
	log.Printf("onerror, name=%s, msg=%s\n", name, err.Error())
}

func (c *Handler) OnPing(socket *gws.Conn, payload []byte) {
	_ = socket.SetDeadline(time.Now().Add(PingInterval + PingWait))
	_ = socket.WritePong([]byte("pong"))
}

func (h *Handler) Broadcast(conn *gws.Conn, opcode gws.Opcode, payload []byte) {
	var b = gws.NewBroadcaster(opcode, payload)
	defer func() {
		_ = b.Close()
	}()
	cname := h.getName(conn)
	ckey := h.getKey(conn)
	h.sessions.Range(func(key string, value *gws.Conn) bool {
		name := h.getName(value)
		if key != ckey && name != cname {
			b.Broadcast(value)
		}
		return true
	})
}

func (c *Handler) OnPong(socket *gws.Conn, payload []byte) {
	fmt.Println("onPong\n")
}

type MsgOpcode uint8

const (
	ChangeMsgId MsgOpcode = 0x0
	RecieveMsg  MsgOpcode = 0x1
)

type SocketMsgData struct {
	OpCode MsgOpcode   `json:"opCode"`
	Data   interface{} `json:"data"`
}

type ChangeId struct {
	UUID string `json:"uuid"`
	Id   int64  `json:"id"`
}

type MyMessage struct {
	db.Message
	UUID string `json:"uuid"`
}

func (c *Handler) OnMessage(socket *gws.Conn, message *gws.Message) {
	if b := message.Data.Bytes(); len(b) == 4 && strings.ToLower(string(b)) == "ping" {
		defer func() {
			err := message.Close()
			if err != nil {
				fmt.Println("关闭消息失败:" + err.Error())
			}
		}()
		c.OnPing(socket, []byte("pong"))
		return
	}
	fmt.Println("收到消息:" + message.Data.String())
	// 保存消息
	m := &MyMessage{}
	err := json.Unmarshal(message.Data.Bytes(), m)
	if err != nil {
		return
	}
	db.MsgDB.CreateMessage(&m.Message)
	socketMsgData := SocketMsgData{
		OpCode: RecieveMsg,
		Data:   m,
	}
	if c.replySend(socket, m, err) {
		return
	}
	broadcastContents, err := json.Marshal(socketMsgData)
	if err != nil {
		fmt.Println("消息序列化失败" + err.Error())
		return
	}
	defer func() {
		_ = message.Close()
	}()
	c.Broadcast(socket, message.Opcode, broadcastContents)
}

func (c *Handler) replySend(socket *gws.Conn, m *MyMessage, err error) bool {
	currentSocketReplayData := SocketMsgData{
		OpCode: ChangeMsgId,
		Data: ChangeId{
			UUID: m.UUID,
			Id:   m.Id,
		},
	}
	sendContent, err := json.Marshal(currentSocketReplayData)
	if err != nil {
		return true
	}
	socket.WriteMessage(gws.OpcodeText, sendContent)
	return false
}
