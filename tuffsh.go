package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	tuffshtools "tuffsh/tuffsh-tools"
)

func main() {
	fmt.Println("TuffSH - SSH Client")
	in, out, e := tuffshtools.TuffSSH(tuffshtools.Host, tuffshtools.Port)
	if e != nil {
		log.Fatalf("Error: SSH to %#v failed with %#v", tuffshtools.Host, fmt.Sprintf("%s", e))
	}
	fmt.Printf("Tuff SSH Client connected to %#v:\n", tuffshtools.Host)
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
	fmt.Printf("Tuff SSH session to %#v ended:\n", tuffshtools.Host)
}
