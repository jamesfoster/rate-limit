package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	flag "github.com/ogier/pflag"
)

type unit interface{}

func main() {
	var port *int

	initialRate, port, batchSize, err := parseCommandLine()

	if nil != err {
		fmt.Println()
		fmt.Println(err)
		return
	}

	stream := make(chan string)
	limiter := make(chan unit)
	updater := make(chan float64)

	go readFromStdIn(stream, limiter, batchSize)

	go rateLimit(limiter, updater, initialRate)

	go listenForChangesToRate(port, updater)

	writeToStdOut(stream)
}

func parseCommandLine() (initialRate float64, port *int, batchSize *int, err error) {

	port = flag.Int("port", 0, "port to listen for rate changes")
	batchSize = flag.Int("batchsize", 1, "items in each batch")

	flag.Parse()

	str := flag.Arg(0)

	if str == "" {
		err = errors.New("must specify an initial rate")
		flag.Usage()
		return
	}

	initialRate, err = strconv.ParseFloat(str, 64)

	if nil != err {
		flag.Usage()
	}

	return
}

func readFromStdIn(stream chan string, limiter chan unit, batchSize *int) {
	in := bufio.NewReader(os.Stdin)

	for {
		<-limiter

		for i := 0; i < *batchSize; i++ {
			line, err := in.ReadString('\n')

			stream <- line

			if io.EOF == err {
				close(stream)
				return
			}
		}
	}
}

func rateLimit(limiter chan unit, updater chan float64, initialRate float64) {
	rate := initialRate
	period := int64(math.Ceil(1000 / rate))

	paused := make(chan time.Time)

	for {
		var wait <-chan time.Time

		if rate == 0 {
			wait = paused
		} else {
			wait = time.After(time.Duration(period) * time.Millisecond)
		}

		select {
		case <-wait:
			limiter <- nil
		case rate = <-updater:
			period = int64(math.Ceil(1000 / rate))
		}
	}
}

func writeToStdOut(stream chan string) {
	out := os.Stdout

	for line := range stream {
		out.WriteString(line)
	}
}

func listenForChangesToRate(port *int, updater chan float64) {
	if *port == 0 {
		return
	}

	address := fmt.Sprintf(":%d", *port)

	err := http.ListenAndServe(address, updateRate(updater))

	if err != nil {
		panic(err)
	}
}

func updateRate(updater chan float64) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, 10)

		size, err := r.Body.Read(buf)

		if io.EOF != err {
			http.Error(w, "Error reading rate", http.StatusBadRequest)
			return
		}

		str := string(buf[:size])

		newRate, err := strconv.ParseFloat(str, 64)

		if nil != err {
			http.Error(w, "Error parsing rate", http.StatusBadRequest)
			return
		}

		updater <- newRate

		w.Write([]byte("Rate changed"))
	})
}
