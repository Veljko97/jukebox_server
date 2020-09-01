package music

import (
	"fmt"
	"github.com/Veljko97/jukebox_server/pkg/utils"
	"github.com/kkdai/youtube"
	"github.com/xfrr/goffmpeg/ffmpeg"
	"github.com/xfrr/goffmpeg/transcoder"
	"io"
	"mime/multipart"
	"os"
	"strings"
	"sync"
)

var SongUploadWaitGroup sync.WaitGroup

func DownloadYoutubeSong(songURL string) {
	defer SongUploadWaitGroup.Done()
	client := youtube.NewYoutube(false, false)
	client.DecodeURL(songURL)

	err := client.StartDownload(utils.MainMusicDir,
		client.Title+ "." + utils.Mp4Extension, "", 0)
	fileLocation := convertMp4ToMp3(utils.MainMusicDir + string(os.PathSeparator) + client.Title)
	newSong := NewSongAdded{
		Song: &Song{
			id:            -1,
			Location:      fileLocation,
			Name:          client.Title,
			AudioFileType: utils.Mp3Extension,
		},
		Err: err,
	}

	NewSongChan <- newSong
	SongUploadWaitGroup.Done()
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

func AddSongFile(fileHeader *multipart.FileHeader){
	defer SongUploadWaitGroup.Done()
	tokens := strings.Split(fileHeader.Filename, string(os.PathSeparator))
	filename := tokens[len(tokens) - 1]
	newFileLocation := utils.MainMusicDir + string(os.PathSeparator) + filename
	if songName, songType := utils.FormatSongName(filename); strings.ToLower(songType) == utils.Mp3Extension {
		file, _ := fileHeader.Open()
		defer file.Close()
		out, _ := os.Create(utils.MainMusicDir + string(os.PathSeparator) + filename)
		defer out.Close()

		_, err := io.Copy(out, file)

		if err != nil {
			fmt.Println(err)
			return
		}

		newSong := NewSongAdded{
			Song: &Song{
				id:            -1,
				Location:      newFileLocation,
				Name:          songName,
				AudioFileType: utils.Mp3Extension,
			},
			Err: nil,
		}
		NewSongChan <- newSong
	}
}