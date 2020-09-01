package music

import (
	json2 "encoding/json"
	"fmt"
	"github.com/Veljko97/jukebox_server/pkg/music"
	"github.com/Veljko97/jukebox_server/pkg/utils"
	"github.com/Veljko97/jukebox_server/pkg/websocket"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"strconv"
)

func AddRoutes(){
	musicPrefix := "/music"
	prefix := utils.ApiPrefix + musicPrefix
	lockedPrefix :=  utils.ApiPrefix + utils.ApiLock + musicPrefix
	utils.Router.HandleFunc(prefix + "/getDetails", SongDetails)
	utils.Router.HandleFunc(prefix + "/getSongList", GetSongsList)
	utils.Router.HandleFunc(prefix + "/voteOnSong/{id}", VoteOnSong)
	utils.Router.HandleFunc(prefix + "/getCurrentSong", GetCurrentSong)
	utils.Router.HandleFunc(lockedPrefix + "/playStop", PlayMusic)
	utils.Router.HandleFunc(lockedPrefix + "/nextSong", NextSong).Methods(http.MethodPut)
	utils.Router.HandleFunc(lockedPrefix + "/youTubeDownload", DownloadFromYouTube).Methods(http.MethodPost)
	utils.Router.HandleFunc(lockedPrefix + "/addMusic", AddMusicFile).Methods(http.MethodPost)
}


func PlayMusic(w http.ResponseWriter, r *http.Request) {
	music.StartPauseMusic()
}

func SongDetails(w http.ResponseWriter, r *http.Request) {
	songDetails := music.GetSongData()
	json, _ := json2.Marshal(songDetails)
	_, _ = w.Write(json)
}

func NextSong(w http.ResponseWriter, r *http.Request) {
	music.NextSong()
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(""))
}

func GetSongsList(w http.ResponseWriter, r *http.Request) {
	songVotes := music.GetVotingList()
	json, _ := json2.Marshal(songVotes)
	_, _ = w.Write(json)
}

func VoteOnSong(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	songKey := vars["id"]
	websocket.PingAll()
	if _, err := strconv.Atoi(songKey); err != nil{
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(""))
	}
	songId, _ := strconv.Atoi(songKey)
	userAddress, _ := utils.GetIpAddress(r)
	songVotes := music.VoteOnSong(userAddress, songId)
	websocket.SendObjectToAll(music.NextSongStarted{VotingList: songVotes})
	json, _ := json2.Marshal(songVotes)
	_, _ = w.Write(json)
}

func GetCurrentSong(w http.ResponseWriter, r *http.Request) {
	song := music.GetSongData()
	json, _ := json2.Marshal(song)
	w.Write(json)
}

func DownloadFromYouTube(w http.ResponseWriter, r *http.Request) {
	bytes, _ := ioutil.ReadAll(r.Body)
	var songUrls map[string][]string
	json2.Unmarshal(bytes, &songUrls)

	for _, songUrl := range songUrls["links"] {
		music.SongUploadWaitGroup.Add(1)
		go music.DownloadYoutubeSong(songUrl)
	}
	music.SongUploadWaitGroup.Wait()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(""))
}


func AddMusicFile(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(200000) // grab the multipart form
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	formdata := r.MultipartForm // ok, no problem so far, read the Form data

	//get the *fileheaders
	files := formdata.File["songs"] // grab the filenames

	for _, fileHarder := range files{
		music.SongUploadWaitGroup.Add(1)
		go music.AddSongFile(fileHarder)
	}
	music.SongUploadWaitGroup.Wait()
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(""))
}