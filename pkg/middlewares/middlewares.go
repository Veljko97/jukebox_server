package middlewares

import (
	"encoding/json"
	"fmt"
	"github.com/Veljko97/jukebox_server/pkg/utils"
	"net/http"
	"strings"
)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		defer func() {
			err := recover()
			if err != nil {
				fmt.Println(err) // May be log this error? Send to sentry?

				jsonBody, _ := json.Marshal(map[string]string{
					"error": "There was an internal server error",
				})

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write(jsonBody)
			}

		}()

		next.ServeHTTP(w, r)

	})
}

func HostCheck(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		defer func() {
			serverAddress := *utils.ServerData.ServerAddress + utils.ServerPort
			localHostAddress := utils.LocalHost + utils.LocalHost
			if (serverAddress == r.Host || localHostAddress == r.Host) && strings.HasPrefix(r.RequestURI, "api/") {

				jsonBody, _ := json.Marshal(map[string]string{
					"error": "You cannot access this",
				})

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				w.Write(jsonBody)
				return
			}
		}()

		next.ServeHTTP(w, r)

	})
}
