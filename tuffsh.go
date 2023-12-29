package main

import (
	"fmt"
	tuffshtools "tuffsh/tuffsh-tools"
)

func main() {
	e := tuffshtools.CheckArgs()
	if e != nil {
		return
	}
	e = tuffshtools.CreateSession()
	if e != nil {
		fmt.Printf("\n%s", e)
	}
}
