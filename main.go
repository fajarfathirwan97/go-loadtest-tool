package main

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"loadtest-tool/loadtest"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func startServer() {
	r := mux.NewRouter()
	r.HandleFunc("/", loadtest.LoadTestViewHandler).Methods("GET")
	r.HandleFunc("/ws", loadtest.LoadTestWebsocketHandler).Methods("GET")
	s := &http.Server{
		Handler:      r,
		Addr:         ":5000",
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}
	go func() {
		err := s.ListenAndServe()
		logrus.Println("Web server running on port :5000")
		if err != nil {
			logrus.Fatal(err)
		}
	}()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	s.Shutdown(ctx)
	os.Exit(0)
}
func main() {
	startServer()
}
