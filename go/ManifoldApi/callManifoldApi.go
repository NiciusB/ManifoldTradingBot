package ManifoldApi

import (
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type apiReqType struct {
	method           string
	url              string
	reqBody          io.Reader
	responseCallback func(string)
}

var apiReqQueue []apiReqType
var apiReqQueueLock sync.Mutex
var queueStartLock sync.Mutex

func callManifoldApi(method string, path string, reqBody io.Reader) string {
	if queueStartLock.TryLock() {
		go startConsumingQueue()
	}

	respChan := make(chan string)

	apiReqQueueLock.Lock()
	apiReqQueue = append(apiReqQueue, apiReqType{
		method:  method,
		url:     "https://api.manifold.markets/" + path,
		reqBody: reqBody,
		responseCallback: func(response string) {
			respChan <- response
		},
	})
	apiReqQueueLock.Unlock()

	return <-respChan
}

func startConsumingQueue() {
	var reqPerSecond int64 = 99 // Limit is 100 per second, use 99 to make sure we don't go over
	var tickDuration = time.Duration(time.Second.Nanoseconds() / reqPerSecond)
	ticker := time.NewTicker(tickDuration)
	for {
		<-ticker.C

		apiReqQueueLock.Lock()
		if len(apiReqQueue) > 0 {
			apiReq := apiReqQueue[len(apiReqQueue)-1]
			apiReqQueue = apiReqQueue[:len(apiReqQueue)-1]
			go _consumeQueueApiReq(apiReq)
		}
		apiReqQueueLock.Unlock()
	}
}

func _consumeQueueApiReq(apiReq apiReqType) {
	var debug = os.Getenv("MANIFOLD_API_DEBUG") == "true"
	if debug {
		log.Printf("[callManifoldApi] method: %+v url: %+v body: %+v\n", apiReq.method, apiReq.url, apiReq.reqBody)
	}

	req, err := http.NewRequest(apiReq.method, apiReq.url, apiReq.reqBody)
	if err != nil {
		log.Fatalln(err)
	}

	var manifoldApiKey = os.Getenv("MANIFOLD_API_KEY")

	req.Header.Add("User-Agent", "ManifoldTradingBot/1.0.0 for @NiciusBot")
	req.Header.Add("Authorization", "Key "+manifoldApiKey)
	if apiReq.method == "POST" && apiReq.reqBody != nil {
		req.Header.Add("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalln(err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	sb := string(body)

	apiReq.responseCallback(sb)
}

func GetQueueLength() int {
	return len(apiReqQueue)
}
