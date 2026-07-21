//go:build !solution

package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func main() {
	urls := os.Args[1:]
	for _, url := range urls {
		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("fetch: %v", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(string(body))
	}
}
