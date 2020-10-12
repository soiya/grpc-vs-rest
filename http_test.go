package main

import (
	"crypto/tls"
	"encoding/json"
	"github.com/Bimde/grpc-vs-rest/pb"
	"golang.org/x/net/http2"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"sync"
	"testing"
)

var client http.Client

func init() {
	client = http.Client{}
}


func get(path string, output interface{}) error {
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		log.Println("error creating request ", err)
		return err
	}

	res, err := client.Do(req)
	if err != nil {
		log.Println("error executing request ", err)
		return err
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println("error reading response body ", err)
		return err
	}

	err = json.Unmarshal(bytes, output)
	if err != nil {
		log.Println("error unmarshalling response ", err)
		return err
	}

	return nil
}

type Request struct {
	Path   string
	Random *pb.Random
}

const stopRequestPath = "STOP"
const noWorkers = 128

func BenchmarkHTTP2GetWithWokers(b *testing.B) {
	client.Transport = &http2.Transport{
		// So http2.Transport doesn't complain the URL scheme isn't 'https'
		AllowHTTP: true,
		// Pretend we are dialing a TLS endpoint. (Note, we ignore the passed tls.Config)
		DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
			return net.Dial(network, addr)
		},
	}
	requestQueue := make(chan Request)
	defer startWorkers(&requestQueue, noWorkers, startWorker)()
	b.ResetTimer() // don't count worker initialization time
	for i := 0; i < b.N; i++ {
		requestQueue <- Request{Path: "http://localhost:8080", Random: &pb.Random{}}
	}
}

func BenchmarkHTTP11Get(b *testing.B) {
	client.Transport = &http.Transport{}
	requestQueue := make(chan Request)
	defer startWorkers(&requestQueue, noWorkers, startWorker)()
	b.ResetTimer() // don't count worker initialization time
	for i := 0; i < b.N; i++ {
		requestQueue <- Request{Path: "http://localhost:8080", Random: &pb.Random{}}
	}
}

func startWorkers(requestQueue *chan Request, noWorkers int, startWorker func(*chan Request, *sync.WaitGroup)) func() {
	var wg sync.WaitGroup
	for i := 0; i < noWorkers; i++ {
		startWorker(requestQueue, &wg)
	}
	return func() {
		wg.Add(noWorkers)
		stopRequest := Request{Path: stopRequestPath}
		for i := 0; i < noWorkers; i++ {
			*requestQueue <- stopRequest
		}
		wg.Wait()
	}
}

func startWorker(requestQueue *chan Request, wg *sync.WaitGroup) {
	go func() {
		for {
			request := <-*requestQueue
			if request.Path == stopRequestPath {
				wg.Done()
				return
			}
			get(request.Path, request.Random)
		}
	}()
}
