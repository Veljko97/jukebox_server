package music

import (
	"fmt"
	"github.com/Veljko97/jukebox_server/pkg/utils"
	"github.com/kkdai/youtube"
	"github.com/xfrr/goffmpeg/ffmpeg"
	"github.com/xfrr/goffmpeg/transcoder"
	"os"
)

func DownloadYoutubeSong(songURL string, addSongChan chan NewSongAdded) {
	client := youtube.NewYoutube(false, false)
	client.DecodeURL(songURL)

	err := client.StartDownload(utils.MusicDirectory+utils.MainMusicDir,
		client.Title+ "." + utils.Mp4Extension, "", 0)
	fileLocation := convertMp4ToMp3(utils.MusicDirectory + utils.MainMusicDir + string(os.PathSeparator) + client.Title)
	newSong := NewSongAdded{
		Song: &Song{
			id:            -1,
			Location:      fileLocation,
			Name:          client.Title,
			AudioFileType: utils.Mp3Extension,
		},
		Err: err,
	}

	addSongChan <- newSong
	NewSongChan <- newSong

	//vid, _ := youtube.Get("")
	//vid.Download()
}


func convertMp4ToMp3(fileLocation string) string {
	trans := new(transcoder.Transcoder)
	config := &ffmpeg.Configuration{
		FfmpegBin:  utils.FFmpegBin,
		FfprobeBin: utils.FFprobeBin,
	}
	trans.SetConfiguration(*config)
	// Initialize transcoder passing the input file path and output file path
	err := trans.Initialize( fileLocation + "." + utils.Mp4Extension, fileLocation + "." + utils.Mp3Extension )
	if err != nil {
		fmt.Println(err)
		return ""
	}

	// Start transcoder process without checking progress
	done := trans.Run(false)

	// This channel is used to wait for the process to end
	err = <-done
	os.Remove(fileLocation + "." + utils.Mp4Extension)
	return fileLocation + "." + utils.Mp3Extension
}