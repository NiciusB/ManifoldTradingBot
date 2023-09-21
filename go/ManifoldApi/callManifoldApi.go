package ManifoldApi

import (
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"
)

var maxPerSecond = 99

func callManifoldApi(method string, path string, reqBody io.Reader) string {
	return callManifoldApiWithFullUrl("https://manifold.markets/api/"+method, path, reqBody)
}
func callManifoldApiWithFullUrl(method string, url string, reqBody io.Reader) string {
	computeThroughputLimiter()

	var debug = os.Getenv("MANIFOLD_API_DEBUG") == "true"
	if debug {
		log.Printf("callManifoldApi (%+v %+v), body:\n%+v\n", method, url, reqBody)
	}

	req, err := http.NewRequest(method, url, reqBody)
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
func computeThroughputLimiter() {
	throughputLimiter.lock.Lock()

	var second = time.Now().Unix()
	if second != throughputLimiter.lastSecond {
		throughputLimiter.lastSecond = second
		throughputLimiter.calls = 1
	} else {
		throughputLimiter.calls++
		if throughputLimiter.calls > maxPerSecond {
			throughputLimiter.lock.Unlock()
			time.Sleep(time.Millisecond*time.Duration(rand.Int63n(19)) + 1)
			computeThroughputLimiter()
			return
		}
	}

	throughputLimiter.lock.Unlock()
}

func GetThroughputFillPercentage() float64 {
	throughputLimiter.lock.Lock()
	defer throughputLimiter.lock.Unlock()

	var second = time.Now().Unix()
	if second != throughputLimiter.lastSecond {
		return 0
	} else {
		if throughputLimiter.calls >= maxPerSecond {
			return 1
		} else {
			return float64(throughputLimiter.calls) / float64(maxPerSecond)
		}
	}
}
