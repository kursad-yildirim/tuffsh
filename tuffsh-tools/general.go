package tuffshtools

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"runtime"
	"strings"
)

func CheckArgs() error {
	// Read flags and command line arguments
	userHome, _ := os.UserHomeDir()
	flag.StringVar(&d.UserKey, "i", userHome+"/"+defaultPrivateKeyFile, "User private key path")
	flag.StringVar(&d.UserKey, "identity", userHome+"/"+defaultPrivateKeyFile, "User private key path")
	flag.StringVar(&d.HostKey, "k", userHome+"/"+defaultKnownHostsFile, "Known hosts path")
	flag.StringVar(&d.HostKey, "known-hosts-file", userHome+"/"+defaultKnownHostsFile, "Known hosts path")
	flag.StringVar(&d.Port, "p", defaultSshPort, "SSH Port")
	flag.StringVar(&d.Port, "port", defaultSshPort, "SSH Port")
	flag.BoolVar(&help, "h", false, "Print usage")
	flag.BoolVar(&help, "help", false, "Print usage")
	flag.BoolVar(&version, "v", false, "Print version")
	flag.BoolVar(&version, "version", false, "Print version")
	flag.Parse()
	// Process arguments
	switch {
	case help:
		printUsage()
		return fmt.Errorf("help requested")
	case version:
		printVersion()
		return fmt.Errorf("version requested")
	case len(flag.Args()) == 1:
		return nil
	case len(flag.Args()) >= 2:
		sshCommand = flag.Args()[1]
		for i := 2; i < len(flag.Args()); i++ {
			sshCommand = sshCommand + " " + flag.Args()[i]
		}
		return nil
	default:
		printUsage()
		return fmt.Errorf("error: you must specify destination ssh server")
	}
}

func printUsage() {
	fmt.Printf("usage: tuffsh\n")
	fmt.Printf("\t[-i/--identity identity_file]\n")
	fmt.Printf("\t[-k/--known-hosts known_hosts_file]\n")
	fmt.Printf("\t[-p/--port ssh port number]\n")
	fmt.Printf("\t[-v/--version]\n")
	fmt.Printf("\t[-h/--help]\n")
	fmt.Printf("\t[user@]destination[:port]\n")
	fmt.Printf("\t[command]\n")
}

func printVersion() {
	fmt.Printf("%v\n", versionStr)
}

func (d *destination) getPort() error {
	darray := strings.Split(flag.Args()[0], ":")
	switch {
	case len(darray) == 1:
		return nil
	case len(darray) == 2:
		d.Port = darray[1]
		return nil
	default:
		return fmt.Errorf("error: destination must be in format [user@]destination[:port]")
	}
}

func (d *destination) getUser() error {
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

func (d *destination) getHost() error {
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
