package music

import (
	"fmt"
	"github.com/Veljko97/jukebox_server/pkg/utils"
	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
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

var TempSong []*Song

var songVotesLock sync.Mutex
var SongVotes = make(map[*Song][]string)

var songDone = make(chan bool)

var currentSong PlayingSong

var NewSongStarted = make(chan NextSongStarted)

var songIdMux = sync.Mutex{}
var LastSongId = 0

var NewSongChan = make(chan NewSongAdded)

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
	err := filepath.Walk(utils.MusicDirectory, func(path string, info os.FileInfo, err error) error {
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
		newSong := <- NewSongChan
		if newSong.Err != nil {
			fmt.Println(newSong.Err)
			continue
		}
		songIdMux.Lock()
		LastSongId ++
		newSong.Song.id = LastSongId
		AllSongs = append(AllSongs, newSong.Song)
		songIdMux.Unlock()
	}
}

func prepareNextSet(lastSong *Song) {
	for i := 0; i < utils.NumberOfSongs; {
		randSong := randomSong()
		if lastSong != nil {
			if randSong.id == lastSong.id {
				continue
			}
		}
		if _, ok := SongVotes[randSong]; ok {
			continue
		}
		SongVotes[randSong] = make([]string, 0)
		i++
	}
}

func randomSong() *Song {
	return AllSongs[rand.Intn(len(AllSongs))]
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
	volume := &effects.Volume{
		Streamer: ctrl,
		Base:     2,
		Volume:   0,
		Silent:   true,
	}
	currentSong = PlayingSong{AudioStream: audioStream, AudioFormat: audioFormat, AudioControl: ctrl, SongDetails: song}
	speaker.Play(volume)
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
			speaker.Lock()
			fmt.Println(audioFormat.SampleRate.D(audioStream.Position()).Round(time.Second))
			speaker.Unlock()
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
	songs := make([]VotingSong, utils.NumberOfSongs)
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
				return
			}
		}
	}
}
