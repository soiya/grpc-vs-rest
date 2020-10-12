package main

import (
	"encoding/json"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"log"
	"net/http"

	"github.com/Bimde/grpc-vs-rest/pb"
)

func handle(w http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	var random pb.Random
	if err := decoder.Decode(&random); err != nil {
		panic(err)
	}
	random.RandomString = "[Updated] " + random.RandomString

	bytes, err := json.Marshal(&random)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

func main() {
	h2s := &http2.Server{}
	server := &http.Server{Addr: "localhost:8080", Handler: h2c.NewHandler(http.HandlerFunc(handle), h2s)}
	log.Fatal(server.ListenAndServe())
}
