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
	var d tuff.Destination
	d, e := tuff.CheckArgs()
	if e != nil {
		//fmt.Println(e)
		return
	}
	in, out, e := tuffshtools.TuffSSH(d)
	if e != nil {
		log.Fatalf("Error: SSH to %#v failed with %#v", d.Host, fmt.Sprintf("%s", e))
	}
	fmt.Printf("Tuff SSH Client connected to %#v:\n", d.Host)
	reader := bufio.NewScanner(os.Stdin)
	fmt.Printf("%s", <-out)
	for reader.Scan() {
		if reader.Text() == "exit" {
			break
		}
		in <- reader.Text()
		go func() {
			for resp := range out {
				fmt.Printf("%s", resp)
			}
		}()
	}
	in <- reader.Text()
	fmt.Printf("Tuff SSH session to %#v ended:\n", d.Host)
}
