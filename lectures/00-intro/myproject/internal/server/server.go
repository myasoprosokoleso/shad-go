package server

import (
	"net/http"

	"github.com/google/uuid"
)

func Start() {
	http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = uuid.New().String()
		_, _ = w.Write([]byte("Hello, World!"))
	}))
}
