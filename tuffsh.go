package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"tuffsh/tuff-tools"
	tuffshtools "tuffsh/tuffsh-tools"
)

func main() {
	e := tuff.CheckArgs()
	if e != nil {
		return
	}
	var wg, mwg sync.WaitGroup
	session, e := tuffshtools.CreateSession(tuff.D)
	if e != nil {
		log.Fatalf("session create  failed: %s", e)
	}
	w, e := session.StdinPipe()
	if e != nil {
		log.Fatalf("StdinPipe error: %s", e)
	}
	r, e := session.StdoutPipe()
	if e != nil {
		log.Fatalf("StdoutPipe error: %s", e)
	}
	if err := session.Start("/bin/bash"); err != nil {
		log.Fatalf("session start failed %#v", err)
	}

	fmt.Printf("Tuff SSH Client connected to %#v.\n", tuff.D.Host)
	mwg.Add(1)
	wg.Add(1)
	go func() {
		reader := bufio.NewScanner(os.Stdin)
		for reader.Scan() {
			if reader.Text() == "exit" {
				break
			}
			wg.Add(1)
			w.Write([]byte(reader.Text() + "\n"))
			wg.Wait()
		}
		w.Write([]byte(reader.Text() + "\n"))
		fmt.Printf("Tuff SSH session to %#v ended.\n", tuff.D.Host)
	}()
	go func() {
		var buffer [32 * 1024]byte
		i := 0
		for {
			i++
			n, e := r.Read(buffer[:])
			if e != nil {
				mwg.Done()
				return
			}
			if string(buffer[n-2:n]) == "\r\n" {
				/*if i == 1 {
					continue
				}*/
				fmt.Printf("%v", string(buffer[:n]))
			}
			if buffer[n-2] == '$' || buffer[n-2] == '#' {
				fmt.Printf("%v", strings.ReplaceAll(string(buffer[:n]), "\r\r\n", ""))
				i = 0
				wg.Done()
			}
		}
	}()
	mwg.Wait()
}
