package music

import (
	"fmt"
	"github.com/Veljko97/jukebox_server/pkg/utils"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"log"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)


var AllSongs []*Song

var TempSongs = make(map[string]*Song)

var songVotesLock sync.Mutex
var SongVotes = make(map[*Song][]string)

var songDone = make(chan bool)

var currentSong PlayingSong

var NewSongStarted = make(chan NextSongStarted)
var SongTimeUpdate = make(chan SongDescription)

var songIdMux = sync.Mutex{}
var LastSongId = 0

var NewSongChan = make(chan *Song)

var NewTempSongChan = make(chan *NewTempSong)

func LoadMusicFiles() {
	if _, err := os.Stat(utils.MusicDirectory); os.IsNotExist(err) {
		err := os.MkdirAll(utils.MusicDirectory, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}

	if _, err := os.Stat(utils.MainMusicDir); os.IsNotExist(err) {
		err := os.MkdirAll(utils.MainMusicDir, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}

	if _, err := os.Stat(utils.TempMusicDir); os.IsNotExist(err) {
		err := os.MkdirAll(utils.TempMusicDir, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}
	LastSongId = 0
	err := filepath.Walk(utils.MainMusicDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			LastSongId++
			song := Song{id: LastSongId, Location: path}

			if songName, songType := utils.FormatSongName(info.Name()); strings.ToUpper(songType) == "MP3" {
				song.Name = songName
				song.AudioFileType = songType

				AllSongs = append(AllSongs, &song)
			}
		}

		return nil
	})
	if err != nil {
		log.Println(err)
	}
	go waitForNewSong()
}

// https://stackoverflow.com/a/33451503
func RemoveTempSongs() error {
	dir := utils.TempMusicDir
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}

func RemoveSong(values []*Song, song *Song) []*Song {
	var index int

	for i, value := range values {
		if value == song {
			index = i
			break
		}
	}

	return append(values[:index], values[index+1:]...)
}

func waitForNewSong(){
	for {

		select {
		case newSong := <- NewSongChan:
			songIdMux.Lock()
			LastSongId ++
			newSong.id = LastSongId
			AllSongs = append(AllSongs, newSong)
			songIdMux.Unlock()
		case newTempSong := <- NewTempSongChan:
			songIdMux.Lock()
			LastSongId ++
			newTempSong.Song.id = LastSongId
			TempSongs[newTempSong.UserAddress] = newTempSong.Song
			songIdMux.Unlock()
		}
	}
}

func prepareNextSet(lastSong *Song) {
	isLimitMoved := false
	numberOfSongs := utils.NumberOfSongs
	if utils.NumberOfSongs >= len(AllSongs){
		isLimitMoved = true
		numberOfSongs = len(AllSongs)
	}
	for i := 0; i < numberOfSongs; i++ {
		randSong := randomSong()
		if lastSong != nil {
			if randSong.id == lastSong.id {
				if isLimitMoved {
					numberOfSongs --
				}
				continue
			}
		}
		if _, ok := SongVotes[randSong]; ok {
			continue
		}
		SongVotes[randSong] = make([]string, 0)
	}

	isLimitMoved = false
	numberOfSongs = utils.NumberOfTempSongs
	if utils.NumberOfTempSongs > len(TempSongs){
		isLimitMoved = true
		numberOfSongs = len(TempSongs)
	}
	for i := 0; i < numberOfSongs; i++ {
		randSong := randomTempSong()
		if lastSong != nil {
			if randSong.id == lastSong.id {
				if isLimitMoved {
					numberOfSongs --
				}
				continue
			}
		}
		if _, ok := SongVotes[randSong]; ok {
			continue
		}
		SongVotes[randSong] = make([]string, 0)
	}
}

func randomSong() *Song {
	return AllSongs[rand.Intn(len(AllSongs))]
}

func randomTempSong() *Song{
	tempSongs := GetTempSongList()
	return tempSongs[rand.Intn(len(tempSongs))]
}


func GetTempSongList() []*Song{
	values := make([]*Song, len(TempSongs))
	i := 0
	for k := range TempSongs {
		values[i] = TempSongs[k]
		i++
	}
	return values
}

func StartNextSong() {
	var nextSong *Song
	maxVotes := math.MinInt8
	if len(SongVotes) == 0 {
		nextSong = randomSong()
	} else {
		for song, votes := range SongVotes {
			if len(votes) > maxVotes {
				maxVotes = len(votes)
				nextSong = song
			}
		}
	}
	SongVotes = make(map[*Song][]string)
	prepareNextSet(nextSong)
	go PlaySong(nextSong)
}

func PlaySong(song *Song) {
	f, err := os.Open(song.Location)
	if err != nil {
		log.Println(err)
		StartNextSong()
		AllSongs = RemoveSong(AllSongs, song)
		return
	}

	defer f.Close()
	audioStream, audioFormat, err := mp3.Decode(f)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer audioStream.Close()

	speaker.Init(audioFormat.SampleRate, audioFormat.SampleRate.N(time.Second/10))

	ctrl := &beep.Ctrl{Paused: false, Streamer: beep.Seq(audioStream, beep.Callback(func() {
		songDone <- true
	}))}
	//volume := &effects.Volume{
	//	Streamer: ctrl,
	//	Base:     2,
	//	Volume:   0,
	//	Silent:   true,
	//}
	currentSong = PlayingSong{AudioStream: audioStream, AudioFormat: audioFormat, AudioControl: ctrl, SongDetails: song}
	speaker.Play(ctrl)
	newSong := NextSongStarted{
		NextSong:   GetSongData(),
		VotingList: GetVotingList(),
	}
	NewSongStarted <- newSong
	for {
		select {
		case <-songDone:
			speaker.Close()
			StartNextSong()
			return
		case <-time.After(time.Second):
			fmt.Println(audioFormat.SampleRate.D(audioStream.Position()).Round(time.Second))
		case <- time.After(time.Minute):
			GetSongData()
		}
	}
	//if <-songDone {
	//	speaker.Close()
	//	StartNextSong()
	//}
}

func StartPauseMusic() {
	if currentSong == (PlayingSong{}) {
		return
	}
	speaker.Lock()
	currentSong.AudioControl.Paused = !currentSong.AudioControl.Paused
	speaker.Unlock()
}

func GetSongData() SongDescription {
	songDescription := SongDescription{}
	songDescription.SongId = currentSong.SongDetails.id
	songDescription.Name = currentSong.SongDetails.Name
	speaker.Lock()
	//songDescription.SampleRate = int(currentSong.AudioFormat.SampleRate)
	//songDescription.SongCurrentMilliseconds = currentSong.AudioFormat.SampleRate.D(currentSong.AudioStream.Position()).Milliseconds()
	//songDescription.SongMaxMilliseconds = currentSong.AudioFormat.SampleRate.D(currentSong.AudioStream.Len()).Milliseconds()
	//songDescription.Timestamp = utils.GetTimeStamp()
	songDescription.SampleRate = int(currentSong.AudioFormat.SampleRate)
	songDescription.SongCurrentSample = currentSong.AudioStream.Position()
	songDescription.SongMaxSample = currentSong.AudioStream.Len()
	songDescription.Timestamp = utils.GetTimeStamp()
	speaker.Unlock()
	return songDescription
}

func NextSong() {
	speaker.Close()
	songDone <- true
}

func GetVotingList() []VotingSong {
	songs := make([]VotingSong, len(SongVotes))
	i := 0
	for song, votes := range SongVotes {
		songVotes := VotingSong{
			SongId:    song.id,
			SongName:  song.Name,
			SongVotes: len(votes),
		}
		songs[i] = songVotes
		i++
	}
	return songs
}

func VoteOnSong(userAddress string, songId int) []VotingSong {

	songVotesLock.Lock()
	defer songVotesLock.Unlock()
	for song, votes := range SongVotes {
		if song.id == songId {
			for _, user := range votes {
				if user == userAddress {
					break
				}
			}
			SongVotes[song] = append(votes, userAddress)
		} else {
			for index, user := range votes {
				if user == userAddress {
					SongVotes[song] = utils.RemoveString(votes, index)
					break
				}
			}
		}
	}
	return GetVotingList()
}


func RemoveVote(userAddress string) {
	songVotesLock.Lock()
	defer songVotesLock.Unlock()
	for _, votes := range SongVotes {
		for index, user := range votes {
			if user == userAddress {
				votes = utils.RemoveString(votes, index)
				delete(TempSongs, userAddress)
				return
			}
		}
	}
}

func GetAllSongs() []*Song{
	return AllSongs
}
