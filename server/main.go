package main

import (
	"encoding/json"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"log"
	"net/http"

	"github.com/Bimde/grpc-vs-rest/pb"
)

func handle(w http.ResponseWriter, r *http.Request) {
	random := pb.Random{RandomString: "a_random_string", RandomInt: 1984}
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
