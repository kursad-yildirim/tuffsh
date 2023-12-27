package tuff

import (
	"fmt"
	"os"
	"os/user"
	"runtime"
	"strings"
)

func CheckArgs() (Destination, error) {
	var d Destination
	if len(os.Args) != 2 {
		printUsage()
		return d, fmt.Errorf("error: you must specify destination ssh server")
	}
	if e := d.get(); e != nil {
		fmt.Println(e)
		return d, e
	}
	fmt.Printf("ssh connection will be established to %#v with user %#v\n", d.Host, d.User)
	return d, nil
}

func printUsage() {
	fmt.Printf("usage: ssh [user@]destination[:port]\n")
}

func (d *Destination) getPort() error {
	darr := strings.Split(os.Args[1], ":")
	switch {
	case len(darr) == 1:
		d.Port = "22"
		return nil
	case len(darr) == 2:
		d.Port = darr[1]
		return nil
	default:
		return fmt.Errorf("error: destination must be in format [user@]destination[:port]")
	}
}

func (d *Destination) getUser() error {
	switch {
	case len(strings.Split(os.Args[1], "@")) == 1:
		if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
			user, e := user.Current()
			if e != nil {
				return e
			}
			d.User = user.Username
			return nil
		} else {
			return fmt.Errorf("error: When OS is not Linux or MacOS, username is required in format user@destination[:port]")
		}
	case len(strings.Split(os.Args[1], "@")) == 2:
		d.User = strings.Split(os.Args[1], "@")[0]
		return nil
	default:
		return fmt.Errorf("error: destination must be in format [user@]destination[:port]")
	}
}

func (d *Destination) get() error {
	if e := d.getPort(); e != nil {
		return e
	}
	if e := d.getUser(); e != nil {
		return e
	}
	if len(strings.Split(os.Args[1], ":")) > 2 {
		return fmt.Errorf("error: destination must be in format [user@]destination[:port]")
	}
	switch {
	case len(strings.Split(strings.Split(os.Args[1], ":")[0], "@")) == 1:
		d.Host = strings.Split(strings.Split(os.Args[1], ":")[0], "@")[0]
		return nil
	case len(strings.Split(strings.Split(os.Args[1], ":")[0], "@")) == 2:
		d.Host = strings.Split(strings.Split(os.Args[1], ":")[0], "@")[1]
		return nil
	default:
		return fmt.Errorf("error: destination must be in format user@destination")
	}
}
