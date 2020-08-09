package music

import (
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

type Song struct {
	id            int
	Location      string
	Name          string
	AudioFileType string
}

type PlayingSong struct {
	SongDetails  Song
	AudioStream  beep.StreamSeekCloser
	AudioFormat  beep.Format
	AudioControl *beep.Ctrl
}

type SongDescription struct {
	Timestamp               int64
	Name                    string
	SongCurrentMilliseconds int64
	SongMaxMilliseconds     int64
}

type VotingSong struct {
	SongId    int
	SongName  string
	SongVotes int
}

var AllSongs []Song

var TempSong []Song

var SongPicks [utils.NumberOfSongs]Song

var songVotesLock sync.Mutex
var SongVotes = make(map[Song][]string)

var songDone = make(chan bool)

var currentSong PlayingSong

func LoadMusicFiles() {
	if _, err := os.Stat(utils.MusicDirectory); os.IsNotExist(err) {
		err := os.MkdirAll(utils.MusicDirectory, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}

	if _, err := os.Stat(utils.MusicDirectory + utils.MainMusicDir); os.IsNotExist(err) {
		err := os.MkdirAll(utils.MusicDirectory+utils.MainMusicDir, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}

	if _, err := os.Stat(utils.MusicDirectory + utils.TempMusicDir); os.IsNotExist(err) {
		err := os.MkdirAll(utils.MusicDirectory+utils.TempMusicDir, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}
	id := 0
	err := filepath.Walk(utils.MusicDirectory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			id++
			song := Song{id: id, Location: path}
			tokens := strings.Split(info.Name(), ".")
			if len(tokens) > 2 {
				song.Name = strings.Join(tokens[:len(tokens)-1], " ")
				song.AudioFileType = tokens[len(tokens)-1]
			}else {
				song.Name = tokens[0]
				song.AudioFileType = tokens[1]
			}
			AllSongs = append(AllSongs, song)

		}

		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

func prepareNextSet() {
	for i := 0; i < utils.NumberOfSongs; {
		randSong := randomSong()
		if _, ok := SongVotes[randSong]; ok {
			continue
		}
		SongVotes[randSong] = make([]string, 0)
		i++
	}
}

func randomSong() Song {
	rand.Seed(time.Now().UnixNano())
	return AllSongs[rand.Intn(len(AllSongs)-1)]
}

func StartNextSong() {
	var nextSong Song
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
	SongVotes = make(map[Song][]string)
	go PlaySong(nextSong)
	prepareNextSet()
}

func PlaySong(song Song) {
	f, err := os.Open(song.Location)
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()
	audioStream, audioFormat, err := mp3.Decode(f)
	if err != nil {
		log.Fatal(err)
	}
	defer audioStream.Close()

	speaker.Init(audioFormat.SampleRate, audioFormat.SampleRate.N(time.Second/10))

	ctrl := &beep.Ctrl{Paused: false, Streamer: beep.Seq(audioStream, beep.Callback(func() {
		songDone <- true
	}))}

	currentSong = PlayingSong{AudioStream: audioStream, AudioFormat: audioFormat, AudioControl: ctrl, SongDetails: song}
	speaker.Play(ctrl)
	playNext := <-songDone
	if playNext {
		StartNextSong()
	}
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
	songDescription.Name = currentSong.SongDetails.Name
	speaker.Lock()
	songDescription.SongCurrentMilliseconds = currentSong.AudioFormat.SampleRate.D(currentSong.AudioStream.Position()).Milliseconds()
	songDescription.SongMaxMilliseconds = currentSong.AudioFormat.SampleRate.D(currentSong.AudioStream.Len()).Milliseconds()
	songDescription.Timestamp = utils.TimeToMilliseconds(time.Now().Round(time.Millisecond))
	speaker.Unlock()
	return songDescription
}

func NextSong() {
	speaker.Clear()
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
