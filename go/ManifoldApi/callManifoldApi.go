package ManifoldApi

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

func callManifoldApi(method string, path string, reqBody io.Reader) string {
	computeThroughputLimiter(99)

	// log.Printf("callManifoldApi (%+v %+v), body:\n%+v\n", method, path, reqBody)

	req, err := http.NewRequest(method, fmt.Sprintf("https://manifold.markets/api/%s", path), reqBody)
	if err != nil {
		log.Fatalln(err)
	}

	var manifoldApiKey = os.Getenv("MANIFOLD_API_KEY")

	req.Header.Add("User-Agent", "ManifoldTradingBot/1.0.0 for @NiciusBot")
	req.Header.Add("Authorization", "Key "+manifoldApiKey)
	if method == "POST" && reqBody != nil {
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

	return sb
}

type throughputLimiterType struct {
	lock       sync.Mutex
	lastSecond int64
	calls      int
}

var throughputLimiter throughputLimiterType

/*
Sleeps if called more often than maxPerSecond
*/
func computeThroughputLimiter(maxPerSecond int) {
	throughputLimiter.lock.Lock()

	var second = time.Now().Unix()
	if second != throughputLimiter.lastSecond {
		throughputLimiter.lastSecond = second
		throughputLimiter.calls = 1
	} else {
		throughputLimiter.calls++
		if throughputLimiter.calls > maxPerSecond {
			throughputLimiter.lock.Unlock()
			time.Sleep(time.Millisecond * 100)
			computeThroughputLimiter(maxPerSecond)
			return
		}
	}

	throughputLimiter.lock.Unlock()
}
