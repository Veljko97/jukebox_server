// main.go
package main

import (
	"bytes"
	json2 "encoding/json"
	"fmt"
	"github.com/Veljko97/jukebox_server/pkg/middlewares"
	"github.com/Veljko97/jukebox_server/pkg/music"
	"github.com/Veljko97/jukebox_server/pkg/utils"
	"github.com/Veljko97/jukebox_server/pkg/websocket"
	"github.com/Veljko97/jukebox_server/src/gui"
	musicController "github.com/Veljko97/jukebox_server/src/music"
	"log"
	"math/rand"
	"net/http"
	"time"
)

func getServerTime(w http.ResponseWriter, r *http.Request) {
	timestamp := utils.TimestampModel{
		Timestamp: utils.GetTimeStamp(),
	}
	json, _ := json2.Marshal(timestamp)
	w.Write(json)
}

func sendNewIpAddress() {
	ipAddress := utils.GetServerIp()
	fmt.Println(ipAddress)
	addressString := ipAddress.String()
	for {
		localKey := utils.RandomString(utils.LocalKeyLength)
		serverAddress := utils.ServerAddress{
			Identifier: nil,
			LocalKey:   localKey,
			Location:   addressString,
		}
		body, _ := json2.Marshal(serverAddress)
		resp, _ := http.Post(utils.RecordServerAddress, utils.JsonContentType, bytes.NewBuffer(body))
		if resp.StatusCode == 200 {
			var identifier utils.ServerIdentifier
			decoder := json2.NewDecoder(resp.Body)
			decoder.Decode(&identifier)
			utils.ServerData.Identifier = identifier.Identifier
			utils.ServerData.LocalKey = &localKey
			utils.ServerData.ServerAddress = &addressString
			utils.ServerData.SaveServerInformation()
			return
		}
	}
}

func updateServerData() {
	ipAddress := utils.GetServerIp()
	fmt.Println(ipAddress)
	addressString := ipAddress.String()
	serverAddress := utils.ServerAddress{
		Identifier: utils.ServerData.Identifier,
		LocalKey:   *utils.ServerData.LocalKey,
		Location:   addressString,
	}
	utils.ServerData.ServerAddress = &addressString
	go utils.ServerData.SaveServerInformation()
	body, _ := json2.Marshal(serverAddress)
	http.Post(utils.RecordServerAddress, utils.JsonContentType, bytes.NewBuffer(body))
}

func initServerData() {
	utils.ServerData.ReadServerInformation()
	if utils.ServerData.Identifier != nil {
		updateServerData()
	} else {
		sendNewIpAddress()
	}
}


func handleRequests() {
	initServerData()
	utils.Router.HandleFunc(utils.ApiPrefix + "/getServerTime", getServerTime)
	utils.Router.Use(middlewares.Recovery)
	utils.Router.Use(middlewares.HostCheck)
	websocket.InitWebSocket()
	musicController.AddRoutes()

	//must be last
	gui.AddRoutes()

	go utils.OpenLink("http://localhost" + utils.ServerPort)
	log.Fatal(http.ListenAndServe(utils.ServerPort, utils.Router))


}

func main() {
	rand.Seed(time.Now().UnixNano())
	music.RemoveTempSongs()
	music.LoadMusicFiles()
	//music.StartNextSong()
	handleRequests()
}
