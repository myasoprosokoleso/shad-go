//go:build !solution

package main

import (
	"bufio"
	"fmt"
	"os"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	files := os.Args[1:]
	c := make(map[string]int)

	for _, file := range files {
		f, err := os.Open(file)
		check(err)
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			c[line]++
		}

		check(scanner.Err())
	}

	for k, v := range c {
		if v > 1 {
			fmt.Printf("%d\t%s\n", v, k)
		}
	}

}
