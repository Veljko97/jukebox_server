// main.go
package main

import (
	"github.com/Veljko97/jukebox_server/pkg/middlewares"
	"github.com/Veljko97/jukebox_server/pkg/music"
	"github.com/Veljko97/jukebox_server/pkg/utils"
	"github.com/Veljko97/jukebox_server/pkg/websocket"
	musicController "github.com/Veljko97/jukebox_server/src/music"
	"log"
	"net/http"
)




func handleRequests() {
	utils.Router.HandleFunc("/websocket", websocket.CreateConnection)
	utils.Router.Use(middlewares.Recovery)
	musicController.AddRoutes()
	log.Fatal(http.ListenAndServe(":8080", utils.Router))
}

func main() {
	music.LoadMusicFiles()
	music.StartNextSong()
	handleRequests()
}
