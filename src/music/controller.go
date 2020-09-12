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
	utils.Router.HandleFunc(prefix + "/getSongList", GetSongsList)
	utils.Router.HandleFunc(prefix + "/voteOnSong/{id}", VoteOnSong).Methods(http.MethodPut)
	utils.Router.HandleFunc(prefix + "/getCurrentSong", GetCurrentSong)
	utils.Router.HandleFunc(prefix + "/getCurrentSongList", GetCurrentSongList)
	utils.Router.HandleFunc(prefix + "/getAllSongs", GetAllSong)
	utils.Router.HandleFunc(prefix + "/getTempSongs", GetAllTempSongs)
	utils.Router.HandleFunc(prefix + "/addUserSong", AddTempMusic).Methods(http.MethodPost)
	utils.Router.HandleFunc(lockedPrefix + "/playStop", PlayMusic)
	utils.Router.HandleFunc(lockedPrefix + "/nextSong", NextSong).Methods(http.MethodPut)
	utils.Router.HandleFunc(lockedPrefix + "/youTubeDownload", DownloadFromYouTube).Methods(http.MethodPost)
	utils.Router.HandleFunc(lockedPrefix + "/addMusic", AddMusicFile).Methods(http.MethodPost)
	utils.Router.HandleFunc(prefix + "/getServerKey", GetServerKey).Methods(http.MethodGet)
}


func PlayMusic(w http.ResponseWriter, r *http.Request) {
	music.StartPauseMusic()
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
	if songId, err := strconv.Atoi(songKey); err == nil{
		userAddress, _ := utils.GetIpAddress(r)
		songVotes := music.VoteOnSong(userAddress, songId)
		websocket.SendObjectToAll(
			music.NextSongStarted{
				VotingList: songVotes,
				NextSong: music.SongDescription{ SongId: -1}})
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(""))
		return
	}
	w.WriteHeader(http.StatusBadRequest)
	_, _ = w.Write([]byte(""))
}

func GetCurrentSong(w http.ResponseWriter, r *http.Request) {
	song := music.GetSongData()
	json, _ := json2.Marshal(song)
	w.Write(json)
}

func GetCurrentSongList(w http.ResponseWriter, r *http.Request) {
	song := music.GetVotingList()
	json, _ := json2.Marshal(song)
	w.Write(json)
}

func DownloadFromYouTube(w http.ResponseWriter, r *http.Request) {
	bytes, _ := ioutil.ReadAll(r.Body)
	var songUrls map[string][]string
	json2.Unmarshal(bytes, &songUrls)

	counter := len(songUrls["links"])
	downloadChan := make(chan *music.SongError, counter)

	for _, songUrl := range songUrls["links"] {
		go music.DownloadYoutubeSong(songUrl, downloadChan)
	}

	returningValue := make(map [string]string)
	for ; counter <= 0; counter -- {
		songDone := <- downloadChan
		if songDone.Err != nil{
			returningValue[songDone.SongName] = songDone.Err.Error()
		} else {
			returningValue[songDone.SongName] = ""
		}
		if counter <= 0 {
			break
		}
	}

	w.WriteHeader(http.StatusOK)
	json, _ :=json2.Marshal(returningValue)
	w.Write(json)
}


func AddMusicFile(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(200000) // grab the multipart form
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	formdata := r.MultipartForm
	files := formdata.File["songs"]

	counter := len(files)
	fileChan := make(chan *music.SongError, counter)

	for _, fileHarder := range files{
		go music.AddSongFile(fileHarder, fileChan)
	}

	returningValue := make(map [string]string)
	for ; counter > 0; counter -- {
		songDone := <- fileChan
		if songDone.Err != nil{
			returningValue[songDone.SongName] = songDone.Err.Error()
		} else {
			returningValue[songDone.SongName] = ""
		}
	}

	w.WriteHeader(http.StatusOK)
	json, _ :=json2.Marshal(returningValue)
	w.Write(json)
}

func GetServerKey(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json, _ :=json2.Marshal(map[string]string{"serverKey": utils.AppPrefix + *utils.ServerData.LocalKey})
	w.Write(json)
}

func AddTempMusic(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(200000) // grab the multipart form
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	formdata := r.MultipartForm

	userAddress, _ := utils.GetIpAddress(r)

	returningValue := music.AddTempSongFile(formdata.File["song"][0], userAddress)


	w.WriteHeader(http.StatusOK)
	json, _ :=json2.Marshal(returningValue)
	w.Write(json)
}

func GetAllSong(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json, _ :=json2.Marshal(music.SendSongsModel{Songs: music.GetAllSongs()})
	w.Write(json)
}

func GetAllTempSongs(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json, _ :=json2.Marshal(music.SendSongsModel{Songs: music.GetTempSongList()})
	w.Write(json)
}

