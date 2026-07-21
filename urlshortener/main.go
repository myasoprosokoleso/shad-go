//go:build !solution

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
)

type urlShortener struct {
	port string
	keys *concurrentMap[string, string]
	mux  *http.ServeMux
}

func newURLShortener(port string) *urlShortener {
	mux := http.NewServeMux()
	us := &urlShortener{
		port: port,
		keys: newConcurrentMap[string, string](),
		mux:  mux,
	}

	mux.HandleFunc("POST /shorten", us.handleShorten)
	mux.HandleFunc("GET /go/{key}", us.handleGo)
	return us
}

func (us *urlShortener) handleShorten(w http.ResponseWriter, r *http.Request) {
	var req struct{ URL string }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	key := generateKey(req.URL)
	us.keys.Set(key, req.URL)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(struct {
		URL string `json:"url"`
		Key string `json:"key"`
	}{URL: req.URL, Key: key}); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (us *urlShortener) handleGo(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	url, ok := us.keys.Get(key)
	if !ok {
		http.Error(w, "key not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, url, http.StatusFound)
}

func (us *urlShortener) run() {
	log.Printf("Starting server on %s port\n", us.port)

	if err := http.ListenAndServe(us.port, us.mux); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}

func main() {
	port := flag.Int("port", 8080, "port to listen on")
	flag.Parse()

	urlShortener := newURLShortener(fmt.Sprintf(":%d", *port))
	urlShortener.run()
}
