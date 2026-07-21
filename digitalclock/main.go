//go:build !solution

package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	digitWidth  = 8
	digitHeight = 12
	colonWidth  = 4
)

var matrices = make(map[byte][][]bool, 11)

type clockImage struct {
	img           *image.RGBA
	width, height int
	scale         int
	timeStr       string
}

func newClockImage(timeStr string, scale int) *clockImage {
	width := (6*digitWidth + 2*colonWidth) * scale
	height := digitHeight * scale
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	return &clockImage{
		img:     img,
		width:   width,
		height:  height,
		scale:   scale,
		timeStr: timeStr,
	}
}

func (ci *clockImage) render(w io.Writer) error {
	ci.fillBackground()

	var offsetX int
	for i := range ci.timeStr {
		digit := ci.timeStr[i]
		ci.drawDigit(digit, offsetX)
		if digit == ':' {
			offsetX += colonWidth * ci.scale
		} else {
			offsetX += digitWidth * ci.scale
		}
	}

	return png.Encode(w, ci.img)
}

func (ci *clockImage) fillBackground() {
	for y := range ci.height {
		for x := range ci.width {
			ci.img.SetRGBA(x, y, color.RGBA{255, 255, 255, 255}) // white background
		}
	}
}

func (ci *clockImage) drawDigit(digit byte, offsetX int) {
	matrix := digitToMatrix(digit)
	if matrix == nil {
		return
	}

	for y, row := range matrix {
		for x, ok := range row {
			if !ok {
				continue
			}

			for dy := range ci.scale {
				for dx := range ci.scale {
					ci.img.SetRGBA(offsetX+x*ci.scale+dx, y*ci.scale+dy, Cyan)
				}
			}
		}
	}
}

func main() {
	port := flag.Int("port", 8080, "port to listen on")
	flag.Parse()

	log.Printf("Starting server on %d port\n", *port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), http.HandlerFunc(handleClock)); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}

func handleClock(w http.ResponseWriter, r *http.Request) {
	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		http.Error(w, "invalid query parameters", http.StatusBadRequest)
		return
	}

	var (
		timeParam string
		kParam    int
	)
	if len(params["time"]) == 0 || params["time"][0] == "" {
		timeParam = time.Now().Format("15:04:05")
	} else {
		userTime, err := time.Parse("15:04:05", params["time"][0])
		if err != nil || userTime.Format("15:04:05") != params["time"][0] {
			http.Error(w, "invalid time (want 15:04:05 format)", http.StatusBadRequest)
			return
		}
		timeParam = userTime.Format("15:04:05")
	}
	if len(params["k"]) == 0 || params["k"][0] == "" {
		kParam = 1
	} else {
		kParam, err = strconv.Atoi(params["k"][0])
		if err != nil || kParam < 1 || kParam > 30 {
			http.Error(w, "invalid k (want integer between 1 and 30)", http.StatusBadRequest)
			return
		}
	}

	w.Header().Set("Content-Type", "image/png")
	clock := newClockImage(timeParam, kParam)
	if err := clock.render(w); err != nil {
		http.Error(w, "failed to render image", http.StatusInternalServerError)
		return
	}
}

func digitToMatrix(digit byte) [][]bool {
	var pattern string
	switch digit {
	case '0':
		pattern = Zero
	case '1':
		pattern = One
	case '2':
		pattern = Two
	case '3':
		pattern = Three
	case '4':
		pattern = Four
	case '5':
		pattern = Five
	case '6':
		pattern = Six
	case '7':
		pattern = Seven
	case '8':
		pattern = Eight
	case '9':
		pattern = Nine
	case ':':
		pattern = Colon
	default:
		return nil
	}

	if matrix, ok := matrices[digit]; ok {
		return matrix
	}
	matrix := patternToMatrix(pattern)
	matrices[digit] = matrix
	return matrix
}

func patternToMatrix(pattern string) [][]bool {
	lines := strings.Split(pattern, "\n")
	matrix := make([][]bool, len(lines))
	for i, line := range lines {
		matrix[i] = make([]bool, len(line))
		for j, char := range line {
			matrix[i][j] = char == '1'
		}
	}
	return matrix
}
