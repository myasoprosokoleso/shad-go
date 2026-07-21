//go:build !solution

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"go.uber.org/zap"
)

type record struct {
	Athlete                     string
	Country                     string
	Year                        json.Number
	Sport                       string
	Gold, Silver, Bronze, Total int
}

type olympicsServer struct {
	mux    *http.ServeMux
	logger *zap.Logger
	data   []record
}

func newOlympicsServer() *olympicsServer {
	srv := &olympicsServer{mux: http.NewServeMux(), logger: zap.Must(zap.NewDevelopment())}

	srv.mux.Handle("GET /athlete-info", Log(srv.logger)(http.HandlerFunc(srv.handleAthleteInfo)))
	srv.mux.Handle("GET /top-athletes-in-sport", Log(srv.logger)(http.HandlerFunc(srv.handleTopAthletes)))
	srv.mux.Handle("GET /top-countries-in-year", Log(srv.logger)(http.HandlerFunc(srv.handleTopCountries)))
	return srv
}

func (srv *olympicsServer) run(port int, dataPath string) {
	defer func() {
		if err := srv.logger.Sync(); err != nil {
			log.Fatalf("failed to sync logger: %v", err)
		}
	}()

	if err := srv.loadData(dataPath); err != nil {
		log.Fatalf("failed to load data: %v", err)
	}

	log.Printf("Listening on %d port\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), srv.mux); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}

func (srv *olympicsServer) loadData(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &srv.data)
}

func main() {
	port := flag.Int("port", 6029, "port to listen on")
	dataPath := flag.String("data", "./olympics/testdata/olympicWinners.json", "path to json file")
	flag.Parse()

	srv := newOlympicsServer()
	srv.run(*port, *dataPath)
}
