package loadtest

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"html/template"
	"loadtest-tool/entity"
	"log"
	"net/http"
	"path"
)

func LoadTestViewHandler(w http.ResponseWriter, r *http.Request) {
	viewPath := path.Join("views", "loadtest", "index.html")
	tmpl, err := template.ParseFiles(viewPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func LoadTestWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	configB64 := r.URL.Query().Get("config")
	loadTestConfig := entity.LoadTestConfig{}
	config, err := base64.StdEncoding.DecodeString(configB64)
	if err != nil {
		log.Fatal("error:", err)
	}
	_ = json.Unmarshal(config, &loadTestConfig)
	if err != nil {
		logrus.Errorln(err)
		return
	}
	stopChan := make(chan bool)
	loadTest := NewGeneralLoadTest(loadTestConfig.Initial, loadTestConfig.Increment, loadTestConfig.Domain, loadTestConfig.Payload, conn, stopChan)
	ctx, _ := context.WithCancel(context.Background())
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			_ = conn.Close()
			stopChan <- true
		}
		newline := []byte{'\n'}
		space := []byte{' '}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		if len(message) > 0 {
			loadTest.ActionContract(ctx, message)
		}
	}
}
