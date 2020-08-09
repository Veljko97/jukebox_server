package utils

import "os"

const NumberOfSongs = 5

var MusicDirectory = "." + string(os.PathSeparator) + "local_jukebox_music"
var MainMusicDir = string(os.PathSeparator) + "main"
var TempMusicDir = string(os.PathSeparator) + "temp"