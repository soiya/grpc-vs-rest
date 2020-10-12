package main

import (
	"bytes"
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

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func BenchmarkHTTP2PostWithWokers(b *testing.B) {
	client.Transport = &http2.Transport{
		// So http2.Transport doesn't complain the URL scheme isn't 'https'
		AllowHTTP: true,
		// Pretend we are dialing a TLS endpoint. (Note, we ignore the passed tls.Config)
		DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
			return net.Dial(network, addr)
		},
	}
	requestQueue := make(chan Request)
	defer startWorkers(&requestQueue, noWorkers, startPostWorker)()
	b.ResetTimer() // don't count worker initialization time
	for i := 0; i < b.N; i++ {
		requestQueue <- Request{
			Path: "http://localhost:8080",
			Random: &pb.Random{
				RandomInt:    2019,
				RandomString: "a_string",
			},
		}
	}
}

func BenchmarkGRPCWithWokers(b *testing.B) {
	conn, err := grpc.Dial("localhost:9090", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Dial failed: %v", err)
	}
	client := pb.NewRandomServiceClient(conn)
	requestQueue := make(chan Request)
	defer startWorkers(&requestQueue, noWorkers, getStartGRPCWorkerFunction(client))()
	b.ResetTimer() // don't count worker initialization time

	for i := 0; i < b.N; i++ {
		requestQueue <- Request{
			Path: "http://localhost:9090",
			Random: &pb.Random{
				RandomInt:    2019,
				RandomString: "a_string",
			},
		}
	}
}

func post(path string, input interface{}, output interface{}) error {
	data, err := json.Marshal(input)
	if err != nil {
		log.Println("error marshalling input ", err)
		return err
	}
	body := bytes.NewBuffer(data)

	req, err := http.NewRequest("POST", path, body)
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

func getStartGRPCWorkerFunction(client pb.RandomServiceClient) func(*chan Request, *sync.WaitGroup) {
	return func(requestQueue *chan Request, wg *sync.WaitGroup) {
		go func() {
			for {
				request := <-*requestQueue
				if request.Path == stopRequestPath {
					wg.Done()
					return
				}
				client.DoSomething(context.TODO(), request.Random)
			}
		}()
	}
}

func startPostWorker(requestQueue *chan Request, wg *sync.WaitGroup) {
	go func() {
		for {
			request := <-*requestQueue
			if request.Path == stopRequestPath {
				wg.Done()
				return
			}
			post(request.Path, request.Random, request.Random)
		}
	}()
}
