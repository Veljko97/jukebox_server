package music

import "github.com/faiface/beep"

type Song struct {
	id            int    `json:"id"`
	Location      string `json:"-"`
	Name          string `json:"name"`
	AudioFileType string `json:"-"`
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

type SongError struct {
	SongName string
	Err      error
}

type NewTempSong struct {
	UserAddress string
	Song        *Song
}

type SendSongsModel struct {
	Songs []*Song `json:"songs"`
}
