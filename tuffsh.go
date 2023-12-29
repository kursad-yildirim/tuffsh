package main

import (
	"log"
	"tuffsh/tuff-tools"
	tuffshtools "tuffsh/tuffsh-tools"
)

func main() {
	e := tuff.CheckArgs()
	if e != nil {
		return
	}
	e = tuffshtools.CreateSession(tuff.D)
	if e != nil {
		log.Fatalf("session create  failed: %s", e)
	}
}
