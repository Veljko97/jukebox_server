package utils

import "os"

const NumberOfSongs = 5

const MusicDirectory = "." + string(os.PathSeparator) + "local_jukebox_music"
const MainMusicDir = string(os.PathSeparator) + "main"
const TempMusicDir = string(os.PathSeparator) + "temp"

const DataDirectory = "." + string(os.PathSeparator) + "data"
const DataFile = DataDirectory + string(os.PathSeparator) + "data.json"

const RecordServerAddress = "http://localhost:5001/local-jubox/us-central1/addServer" //Testing
//const RecordServerAddress ="https://us-central1-local-jubox.cloudfunctions.net/addServer" //Production

const LocalKeyLength = 6

const JsonContentType = "application/json"