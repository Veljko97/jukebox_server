package utils

import (
	json2 "encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type TimestampModel struct {
	Timestamp int64
}

type ServerAddress struct {
	Identifier *string `json:"identifier"`
	LocalKey   string  `json:"localKey"`
	Location   string  `json:"location"`
}

type ServerIdentifier struct {
	Identifier *string `json:"identifier"`
}

type ServerInformation struct {
	Identifier    *string `json:"identifier"`
	LocalKey      *string `json:"localKey"`
	ServerAddress *string `json:"serverAddress"`
}

//var SeededRand =

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var ServerData = ServerInformation{}

func GetIpAddress(r *http.Request) (string, error) {
	//Get IP from the X-REAL-IP header
	ip := r.Header.Get("X-REAL-IP")
	netIP := net.ParseIP(ip)
	if netIP != nil {
		return ip, nil
	}

	//Get IP from X-FORWARDED-FOR header
	ips := r.Header.Get("X-FORWARDED-FOR")
	splitIps := strings.Split(ips, ",")
	for _, ip := range splitIps {
		netIP := net.ParseIP(ip)
		if netIP != nil {
			return ip, nil
		}
	}

	//Get IP from RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", err
	}
	netIP = net.ParseIP(ip)
	if netIP != nil {
		return ip, nil
	}
	return "", fmt.Errorf("No valid ip found")
}

func IsLocalIp(ipAddress string, localIp string) bool{
	return ipAddress == LocalHostAddress || ipAddress == localIp
}

func TimeToMilliseconds(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}

func RemoveString(values []string, index int) []string {
	return append(values[:index], values[index+1:]...)
}

func GetServerIp() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

func GetTimeStamp() int64 {
	return TimeToMilliseconds(time.Now().Round(time.Millisecond))
}

func RandomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func (si ServerInformation) ReadServerInformation() {
	if _, err := os.Stat(DataDirectory); os.IsNotExist(err) {
		err := os.MkdirAll(DataDirectory, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}
	jsonFile, err := os.OpenFile(DataFile, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	json2.Unmarshal(byteValue, &ServerData)
}

func (si ServerInformation) SaveServerInformation() {
	if _, err := os.Stat(DataDirectory); os.IsNotExist(err) {
		err := os.MkdirAll(DataDirectory, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}
	jsonFile, err := os.OpenFile(DataFile, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()

	byteValue, err := json2.Marshal(ServerData)

	if err != nil {
		log.Println(err)
	}
	_, err = jsonFile.Write(byteValue)

	if err != nil {
		log.Println(err)
	}
}

func FormatSongName(songName string) (string, string){
	tokens := strings.Split(songName, ".")
	var name, format string
	if len(tokens) > 2 {
		name = strings.Join(tokens[:len(tokens)-1], " ")
		format = tokens[len(tokens)-1]
	} else {
		name = tokens[0]
		format = tokens[1]
	}
	return name, format
}

// https://stackoverflow.com/a/39324149
func OpenLink(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

