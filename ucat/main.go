package main

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/eapache/channels.v1"
)

func main() {
	ch := channels.NewInfiniteChannel()

	go func() {
		defer ch.Close()

		for {
			buf := make([]byte, os.Getpagesize())

			n, err := os.Stdin.Read(buf)
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}

			ch.In() <- buf[:n]
		}
	}()

	for m := range ch.Out() {
		buf := m.([]byte)

		for len(buf) != 0 {
			n, err := os.Stdout.Write(buf)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}

			buf = buf[n:]
		}
	}
}
