package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/schema"
	_ "github.com/zuercher/slackbot/importer"
	"github.com/zuercher/slackbot/robots"
)

func Main(robotMap map[string][]robots.Robot) {
	log.Println("[Info] Starting up with the following Robots: ")
	for key, _ := range robotMap {
		log.Println("  ", key)
	}
	http.HandleFunc("/slack", slashCommandHandler(robotMap))
	http.HandleFunc("/slack_hook", hookHandler(robotMap))
	startServer()
}

func hookHandler(robotMap map[string][]robots.Robot) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		d := schema.NewDecoder()
		command := new(robots.OutgoingWebHook)
		err = d.Decode(command, r.PostForm)
		if err != nil {
			log.Println("Couldn't parse post request:", err)
		}
		if command.Text == "" || command.Token != getOutToken(command.TeamDomain) {
			log.Printf("[DEBUG] Ignoring request from unidentified source: %s - %s - %s", command.Token, r.Host, command.TeamDomain)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		com := strings.TrimPrefix(command.Text, command.TriggerWord)
		c := strings.Split(com, " ")
		command.Robot = c[0]
		command.Text = strings.Join(c[1:], " ")

		robots := robots.Robots[command.Robot]
		if len(robots) == 0 {
			jsonResp(w, "No robot for that command yet :(")
			return
		}
		resp := ""
		for _, robot := range robots {
			resp += fmt.Sprintf("\n%s", robot.Run(&command.Payload))
		}
		w.WriteHeader(http.StatusOK)
		jsonResp(w, strings.TrimSpace(resp))
	}
}

func slashCommandHandler(robotMap map[string][]robots.Robot) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		d := schema.NewDecoder()
		command := new(robots.SlashCommand)
		err = d.Decode(command, r.PostForm)
		if err != nil {
			log.Println("Couldn't parse post request:", err)
		}
		if command.Command == "" || command.Token == "" {
			log.Printf("[DEBUG] Ignoring request from unidentified source: %s - %s", command.Token, r.Host)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		command.Robot = command.Command[1:]

		if token := getSlackToken(command.Robot); token != "" && token != command.Token {
			log.Printf("[DEBUG] Ignoring request from unidentified source: %s - %s", command.Token, r.Host)
			w.WriteHeader(http.StatusBadRequest)
		}
		robots := robots.Robots[command.Robot]
		if len(robots) == 0 {
			plainResp(w, "No robot for that command yet :(")
			return
		}
		resp := ""
		for _, robot := range robots {
			resp += fmt.Sprintf("\n%s", robot.Run(&command.Payload))
		}
		plainResp(w, strings.TrimSpace(resp))
	}
}

func jsonResp(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	resp := map[string]string{"text": msg}
	r, err := json.Marshal(resp)
	if err != nil {
		log.Println("Couldn't marshal hook response:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(r)
}

func plainResp(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(msg))
}

func getSlackToken(robot string) string {
	return os.Getenv(fmt.Sprintf("%s_SLACK_TOKEN", strings.ToUpper(robot)))
}

func getOutToken(teamDomain string) string {
	return os.Getenv(fmt.Sprintf("%s_OUT_TOKEN", strings.ToUpper(teamDomain)))
}

func startServer() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT not set")
	}
	log.Printf("Starting HTTP server on %s", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal("Server start error: ", err)
	}
}
