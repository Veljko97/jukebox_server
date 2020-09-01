package websocket

import (
	"github.com/Veljko97/jukebox_server/pkg/music"
	"github.com/Veljko97/jukebox_server/pkg/utils"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

type UserConnection struct {
	Conn *websocket.Conn
	IpAddress string
}

var addressConnection = make(map[string]UserConnection)

var connections []UserConnection

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func removeConnection(index int){
	userAddress := connections[index].IpAddress
	connections = append(connections[:index], connections[index+1:]...)
	delete(addressConnection, userAddress)
}

func InitWebSocket(){

	utils.Router.HandleFunc("/ws/music", createConnection)
	go nextSongSending()
}

func createConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Panic(err)
		return
	}
	userAddress, _ := utils.GetIpAddress(r)
	userConn := UserConnection{Conn: conn, IpAddress: userAddress}
	if oldConnection,ok :=  addressConnection[userAddress]; ok {
		for i, _ := range connections {
			if connections[i] == oldConnection {
				connections[i] = userConn
				break
			}
		}
	}else {
		connections = append(connections, userConn)
	}
	addressConnection[userAddress] = userConn
}

func SendMessageToAll(message string) {
	for index, conn := range connections {
		err := conn.Conn.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			removeConnection(index)
			music.RemoveVote(conn.IpAddress)
			conn.Conn.Close()
		}
	}
}

func SendObjectToAll(obj interface{}) {
	for index, conn := range connections {
		err := conn.Conn.WriteJSON(obj)
		if err != nil {
			removeConnection(index)
			music.RemoveVote(conn.IpAddress)
			conn.Conn.Close()
		}
	}
}

func PingAll(){
	for index, conn := range connections {
		err := conn.Conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(time.Second * 5))
		if err != nil {
			removeConnection(index)
			music.RemoveVote(conn.IpAddress)
			conn.Conn.Close()
		}
	}
}

func nextSongSending(){
	for {
		songDescription := <- music.NewSongStarted
		SendObjectToAll(songDescription)
	}
}
