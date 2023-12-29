package tuff

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"runtime"
	"strings"
)

func CheckArgs() error {
	userHome, _ := os.UserHomeDir()
	flag.StringVar(&D.UserKey, "i", userHome+"/"+defaultPrivateKeyFile, "User private key path")
	flag.StringVar(&D.HostKey, "k", userHome+"/"+defaultKnownHostsFile, "Known hosts path")
	flag.BoolVar(&help, "h", false, "Print usage")
	flag.Parse()
	switch {
	case help:
		printUsage()
		return fmt.Errorf("help requested")
	case len(flag.Args()) != 1:
		printUsage()
		return fmt.Errorf("error: you must specify destination ssh server")
	default:
		if e := D.get(); e != nil {
			fmt.Println(e)
			return e
		}
		return nil
	}
}

func printUsage() {
	fmt.Printf("usage: tuffsh [-i identity_file] [-k known_hosts_file] [user@]destination[:port]\n")
}

func (d *Destination) getPort() error {
	darr := strings.Split(flag.Args()[0], ":")
	switch {
	case len(darr) == 1:
		d.Port = defaultSshPort
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
	case len(strings.Split(flag.Args()[0], "@")) == 1:
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
	case len(strings.Split(flag.Args()[0], "@")) == 2:
		d.User = strings.Split(flag.Args()[0], "@")[0]
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
	if len(strings.Split(flag.Args()[0], ":")) > 2 {
		return fmt.Errorf("error: destination must be in format [user@]destination[:port]")
	}
	switch {
	case len(strings.Split(strings.Split(flag.Args()[0], ":")[0], "@")) == 1:
		d.Host = strings.Split(strings.Split(flag.Args()[0], ":")[0], "@")[0]
		return nil
	case len(strings.Split(strings.Split(flag.Args()[0], ":")[0], "@")) == 2:
		d.Host = strings.Split(strings.Split(flag.Args()[0], ":")[0], "@")[1]
		return nil
	default:
		return fmt.Errorf("error: destination must be in format user@destination")
	}
}
