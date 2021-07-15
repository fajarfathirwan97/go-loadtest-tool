package loadtest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"loadtest-tool/entity"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
)

type LoadTest interface {
	DoLoadTest()
	ActionContract(ctxt context.Context, message []byte)
}

type GeneralLoadTest struct {
	Con               *websocket.Conn
	IsLoadTestStarted bool
	Host              string
	Payload           entity.Payload `json:"payloads"`
	BatchPerRequest   int
	IncrementBatch    int
	SuccessCount      int
	FailureCount      int
	TotalRequestCount int
	StopChan          chan bool
	SlowestTime       float64
	FastestTime       float64
}

func (l *GeneralLoadTest) ActionContract(ctx context.Context, message []byte) {
	messageJson := entity.LoadTestMessage{}
	err := json.Unmarshal(message, &messageJson)
	if err != nil {
		logrus.Errorln(err, "INVALID JSON")
		return
	}
	switch messageJson.Type {
	case "start-loadtest":
		if !l.IsLoadTestStarted {
			if l.Con.WriteJSON(map[string]string{"message": "loadtest-started"}) != nil {
				logrus.Errorln(err)
			}
			go l.DoLoadTest()
		}
		break
	case "stop-loadtest":
		if l.IsLoadTestStarted {
			l.StopChan <- true
			if l.Con.WriteJSON(map[string]string{"message": "loadtest-stopped"}) != nil {
				logrus.Errorln(err)
			}
		}
		break

	default:
		break
	}
}

func (l *GeneralLoadTest) DoLoadTest() {
	l.IsLoadTestStarted = true
	shutdown := make(chan struct{})
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	var wg sync.WaitGroup
	ticker := time.NewTicker(time.Millisecond)
	for {
		select {
		case <-ticker.C:
			start := time.Now()
			for i := 0; i < l.BatchPerRequest; i++ {

				wg.Add(1)
				go func() {
					client := &http.Client{}
					l.FastestTime = 0.0
					l.SlowestTime = 0.0
					for _, request := range l.Payload.Requests {
						body, _ := json.Marshal(request.Body)
						r, err := http.NewRequest(request.Method, fmt.Sprintf("%v%v", l.Host, request.Endpoint), bytes.NewBuffer(body))
						for k, v := range request.Headers {
							r.Header.Set(k, v)
						}
						if err != nil {
							logrus.Errorln(err)
							l.FailureCount++
							continue
						}
						resp, err := client.Do(r)
						if err != nil {
							logrus.Errorln(err)
							l.FailureCount++
							continue
						}
						bodyBytes, err := ioutil.ReadAll(resp.Body)
						if err != nil {
							logrus.Errorln(err)
							l.FailureCount++
							continue
						}
						_ = string(bodyBytes)
						if resp.StatusCode >= 200 && resp.StatusCode < 300 {
							l.SuccessCount++
						} else if resp.StatusCode >= 400 && resp.StatusCode < 600 {
							l.FailureCount++
						}
						resp.Body.Close()
						elapsedRequest := time.Since(start).Seconds()
						if l.SlowestTime == 0 || elapsedRequest > l.SlowestTime {
							l.SlowestTime = elapsedRequest
						}
						if l.FastestTime == 0 || elapsedRequest < l.FastestTime {
							l.FastestTime = elapsedRequest
						}
					}
					wg.Done()
				}()
			}
			wg.Wait()
			elapsedGroup := time.Since(start).Seconds()

			l.TotalRequestCount += l.BatchPerRequest
			l.Con.WriteJSON(entity.LoadTestResult{
				Message:           "loadtest-info",
				SuccessCount:      l.SuccessCount,
				FailureCount:      l.FailureCount,
				TotalRequestCount: l.TotalRequestCount,
				Rps:               l.BatchPerRequest,
				Elapsed:           elapsedGroup,
				FastestTime:       l.FastestTime,
				SlowestTime:       l.SlowestTime,
			})
			l.BatchPerRequest += l.IncrementBatch
			l.SuccessCount = 0
			l.FailureCount = 0
		case <-l.StopChan:
			logrus.Println("RECEIVE STOP CHAN")
			close(shutdown)
			l.IsLoadTestStarted = false
			return
		case <-interrupt:
			logrus.Println("RECEIVE STOP IINTERupt")
			close(shutdown)
			l.IsLoadTestStarted = false
			return

		}

	}
}

func NewGeneralLoadTest(initial int, increment int, host string, payload entity.Payload, conn *websocket.Conn, stopChan chan bool) LoadTest {
	return &GeneralLoadTest{
		Host:            host,
		Payload:         payload,
		BatchPerRequest: initial,
		IncrementBatch:  increment,
		Con:             conn,
		StopChan:        stopChan,
	}
}
