package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"tuffsh/tuff-tools"
	tuffshtools "tuffsh/tuffsh-tools"
)

func main() {
	e := tuff.CheckArgs()
	if e != nil {
		return
	}
	in, out, e := tuffshtools.TuffSSH(tuff.D)
	if e != nil {
		log.Fatalf("Error: SSH to %#v failed with %#v", tuff.D.Host, fmt.Sprintf("%s", e))
	}
	fmt.Printf("Tuff SSH Client connected to %#v.\n", tuff.D.Host)
	reader := bufio.NewScanner(os.Stdin)
	resp := <-out
	fmt.Printf("%s", resp)
	for reader.Scan() {
		if reader.Text() == "exit" {
			break
		}
		in <- reader.Text()
		go func() {
			for resp := range out {
				fmt.Printf("%v", resp)
			}
		}()
	}
	in <- reader.Text()
	fmt.Printf("Tuff SSH session to %#v ended.\n", tuff.D.Host)
}
