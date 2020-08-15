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
	musicController "github.com/Veljko97/jukebox_server/src/music"
	"log"
	"net/http"
)

func getServerTime(w http.ResponseWriter, r *http.Request){
	timestamp := utils.TimestampModel{
		Timestamp: utils.GetTimeStamp(),
	}
	json, _ := json2.Marshal(timestamp)
	w.Write(json)
}

func sendNewIpAddress() {
	ipAddress := utils.GetServerIp()
	fmt.Println(ipAddress)
	for {
		localKey := utils.RandomString(utils.LocalKeyLength)
		serverAddress := utils.ServerAddress{
			Identifier: nil,
			LocalKey:   localKey,
			Location:   ipAddress.String(),
		}
		body, _ := json2.Marshal(serverAddress)
		resp, _ := http.Post(utils.RecordServerAddress, utils.JsonContentType, bytes.NewBuffer(body))
		if resp.StatusCode == 200 {
			var identifier utils.ServerIdentifier
			decoder := json2.NewDecoder(resp.Body)
			decoder.Decode(&identifier)
			utils.ServerData.Identifier = identifier.Identifier
			utils.ServerData.LocalKey = &localKey
			utils.ServerData.SaveServerInformation()
			return
		}
	}
}

func updateServerData(){
	ipAddress := utils.GetServerIp()
	fmt.Println(ipAddress)
	serverAddress := utils.ServerAddress{
		Identifier: utils.ServerData.Identifier,
		LocalKey:   *utils.ServerData.LocalKey,
		Location:   ipAddress.String(),
	}
	body, _ := json2.Marshal(serverAddress)
	http.Post(utils.RecordServerAddress, utils.JsonContentType, bytes.NewBuffer(body))
}

func initServerData(){
	utils.ServerData.ReadServerInformation()
	if utils.ServerData.Identifier != nil {
		updateServerData()
	} else {
		sendNewIpAddress()
	}
}

func handleRequests() {
	initServerData()
	utils.Router.HandleFunc("/api/getServerTime", getServerTime)
	utils.Router.Use(middlewares.Recovery)
	websocket.InitWebSocket()
	musicController.AddRoutes()
	log.Fatal(http.ListenAndServe(":8080", utils.Router))
}

func main() {
	music.LoadMusicFiles()
	music.StartNextSong()
	handleRequests()
}
