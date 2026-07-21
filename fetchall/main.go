//go:build !solution

package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func urlFetch(url string, logs chan<- string) {
	start := time.Now()

	resp, err := http.Get(url)
	if err != nil {
		logs <- fmt.Sprint(err)
		return
	}
	defer resp.Body.Close()

	bodySize, err := io.Copy(io.Discard, resp.Body)
	if err != nil {
		logs <- fmt.Sprint(err)
		return
	}

	logs <- fmt.Sprintf("%.2fs  %7d  %s", time.Since(start).Seconds(), bodySize, url)
}

func main() {
	start := time.Now()
	urls := os.Args[1:]
	logs := make(chan string)

	for _, url := range urls {
		go urlFetch(url, logs)
	}

	for range urls {
		fmt.Println(<-logs)
	}

	fmt.Printf("%.2fs elapsed\n", time.Since(start).Seconds())
}
