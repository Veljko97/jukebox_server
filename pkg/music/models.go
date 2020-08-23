package music

import "github.com/faiface/beep"

type Song struct {
	id            int
	Location      string
	Name          string
	AudioFileType string
}

type PlayingSong struct {
	SongDetails  *Song
	AudioStream  beep.StreamSeekCloser
	AudioFormat  beep.Format
	AudioControl *beep.Ctrl
}

type SongDescription struct {
	SongId            int    `json:"songId"`
	Timestamp         int64  `json:"timestamp"`
	Name              string `json:"name"`
	SongCurrentSample int    `json:"songCurrentSample"`
	SongMaxSample     int    `json:"songMaxSample"`
	SampleRate        int    `json:"sampleRate"`
}

type VotingSong struct {
	SongId    int    `json:"songId"`
	SongName  string `json:"songName"`
	SongVotes int    `json:"songVotes"`
}

type NextSongStarted struct {
	NextSong   SongDescription `json:"nextSong"`
	VotingList []VotingSong    `json:"votingList"`
}

type NewSongAdded struct {
	Song *Song
	Err  error
}
