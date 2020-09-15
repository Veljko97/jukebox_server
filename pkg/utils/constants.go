package utils

import (
	"os"
	"runtime"
)

const LocalHost = "localhost"
const LocalHostAddress = "::1"
const ServerPort = ":8080"

const AppPrefix = "jukebox:///"

const ApiPrefix = "/api"
const ApiLock = "/lock"

const NumberOfSongs = 5
const NumberOfTempSongs = 2

const MusicDirectory = "." + string(os.PathSeparator) + "local_jukebox_music"
const MainMusicDir = MusicDirectory + string(os.PathSeparator) + "main"
const TempMusicDir = MusicDirectory + string(os.PathSeparator) + "temp"

const DataDirectory = "." + string(os.PathSeparator) + "data"
const DataFile = DataDirectory + string(os.PathSeparator) + "data.json"

//const RecordServerAddress = "http://localhost:5001/local-jubox/us-central1/addServer" //Testing
const RecordServerAddress ="https://us-central1-local-jubox.cloudfunctions.net/addServer" //Production

const LocalKeyLength = 6

const JsonContentType = "application/json"

const Mp3Extension = "mp3"
const Mp4Extension = "mp4"

const FFmpegRoot = "." + string(os.PathSeparator) + "FFmpeg" + string(os.PathSeparator) + runtime.GOOS
const FFmpegBin = FFmpegRoot + string(os.PathSeparator) + "ffmpeg"
const FFprobeBin = FFmpegRoot + string(os.PathSeparator) + "ffprobe"
