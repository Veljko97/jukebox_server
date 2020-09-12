package music

import (
	"errors"
	"fmt"
	"github.com/Veljko97/jukebox_server/pkg/utils"
	"github.com/kkdai/youtube/v2"
	"github.com/xfrr/goffmpeg/ffmpeg"
	"github.com/xfrr/goffmpeg/transcoder"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
)


func DownloadYoutubeSong(songURL string, requestChan chan *SongError) {
	var songError *SongError
	defer func (){requestChan <- songError}()
	client := youtube.Client{}

	video, err :=client.GetVideo(songURL)
	if err != nil {
		fmt.Println(err)
		songError = &SongError{
			SongName: songURL,
			Err:      err,
		}
		return
	}

	var resp *http.Response
	for _, format := range video.Formats {
		if !strings.HasPrefix(format.MimeType, "video/mp4") {
			continue
		}
		resp, err = client.GetStream(video, &format)
		if err != nil {
			fmt.Println(err)
			songError = &SongError{
				SongName: video.Title,
				Err:      err,
			}
		} else {
			songError = nil
			break
		}
	}
	if songError != nil {
		return
	}
	defer resp.Body.Close()

	file, err := os.Create(utils.MainMusicDir + string(os.PathSeparator) + video.Title + "." + utils.Mp4Extension)
	if err != nil {
		fmt.Println(err)
		songError = &SongError{
			SongName: video.Title,
			Err:      err,
		}
		return
	}

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		fmt.Println(err)
		songError = &SongError{
			SongName: video.Title,
			Err:      err,
		}
		return
	}

	file.Close()
	fileLocation, err := convertMp4ToMp3(utils.MainMusicDir + string(os.PathSeparator) + video.Title)
	if err != nil {
		fmt.Println(err)
		songError = &SongError{
			SongName: video.Title,
			Err:      err,
		}
		return
	}

	newSong := &Song{
		id:            -1,
		Location:      fileLocation,
		Name:          video.Title,
		AudioFileType: utils.Mp3Extension,
	}

	songError = &SongError{
		SongName: video.Title,
		Err:      nil,
	}

	NewSongChan <- newSong
}


func convertMp4ToMp3(fileLocation string) (string, error) {
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
		return "", err
	}

	// Start transcoder process without checking progress
	done := trans.Run(false)

	// This channel is used to wait for the process to end
	err = <-done
	err = os.Remove(fileLocation + "." + utils.Mp4Extension)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return fileLocation + "." + utils.Mp3Extension, nil
}

func AddSongFile(fileHeader *multipart.FileHeader, requestChan chan *SongError) {
	var songError *SongError
	defer func (){requestChan <- songError}()

	filename := fileHeader.Filename
	newFileLocation := utils.MainMusicDir + string(os.PathSeparator) + filename
	if songName, songType := utils.FormatSongName(filename); strings.ToLower(songType) == utils.Mp3Extension {
		file, err := fileHeader.Open()
		if err != nil {
			fmt.Println(err)
			songError = &SongError{
				SongName: songName,
				Err:      err,
			}
			return
		}
		defer file.Close()
		out, err := os.Create(newFileLocation)
		if err != nil {
			fmt.Println(err)
			songError = &SongError{
				SongName: songName,
				Err:      err,
			}
			return
		}
		defer out.Close()

		_, err = io.Copy(out, file)

		if err != nil {
			fmt.Println(err)
			songError = &SongError{
				SongName: songName,
				Err:      err,
			}
			return
		}

		newSong := &Song{
				id:            -1,
				Location:      newFileLocation,
				Name:          songName,
				AudioFileType: utils.Mp3Extension,
			}
		songError = &SongError{
			SongName: songName,
			Err:      nil,
		}
		NewSongChan <- newSong
	} else {
		songError = &SongError{
			SongName: songName,
			Err:      errors.New("Not a MP3 format"),
		}
		return
	}
}

func AddTempSongFile(fileHeader *multipart.FileHeader, userAddress string) *SongError {
	filename := fileHeader.Filename
	newFileLocation := utils.TempMusicDir + string(os.PathSeparator) + filename
	if songName, songType := utils.FormatSongName(filename); strings.ToLower(songType) == utils.Mp3Extension {
		file, err := fileHeader.Open()
		if err != nil {
			fmt.Println(err)
			return &SongError{
				SongName: songName,
				Err:      err,
			}
		}
		defer file.Close()
		out, err := os.Create(newFileLocation)
		if err != nil {
			fmt.Println(err)
			return &SongError{
				SongName: songName,
				Err:      err,
			}
		}
		defer out.Close()

		_, err = io.Copy(out, file)

		if err != nil {
			fmt.Println(err)
			return &SongError{
				SongName: songName,
				Err:      err,
			}
		}

		newSong := &NewTempSong{
			Song: &Song{
				id:            -1,
				Location:      newFileLocation,
				Name:          songName,
				AudioFileType: utils.Mp3Extension,
			},
			UserAddress: userAddress,
		}
		NewTempSongChan <- newSong
		return &SongError{
			SongName: songName,
			Err:      nil,
		}
	} else {
		return &SongError{
			SongName: songName,
			Err:      errors.New("Not a MP3 format"),
		}
	}
}