package tuffshtools

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/user"
	"runtime"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/term"
)

func prepareSigner(userKeyFile string) (ssh.Signer, error) {
	k, e := os.ReadFile(userKeyFile)
	if e != nil {
		return nil, e
	}
	s, e := ssh.ParsePrivateKey(k)
	if e != nil {
		return nil, e
	}

	return s, nil
}

func verifySshServer(host string, remote net.Addr, key ssh.PublicKey) error {
	reader := bufio.NewReader(os.Stdin)
	var wrongKey *knownhosts.KeyError
	callback, e := knownhosts.New(d.HostKey)
	if e != nil {
		return fmt.Errorf("error:  Known hosts file read error %#v", e)
	}
	e = callback(host, remote, key)
	if e == nil {
		return nil
	} else if errors.As(e, &wrongKey) && len(wrongKey.Want) > 0 {
		return wrongKey
	}

	fmt.Printf("Unknown Host: %s Fingerprint: %s \n", host, ssh.FingerprintSHA256(key))
	fmt.Print("Would you like to add it? type yes or no: ")
	r, e := reader.ReadString('\n')
	if e != nil {
		log.Fatal(e)
	}

	if strings.ToLower(strings.TrimSpace(r)) == "yes" {
		f, e := os.OpenFile(d.HostKey, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if e != nil {
			return e
		}
		defer f.Close()
		_, e = f.WriteString(knownhosts.Line(append([]string{knownhosts.Normalize(remote.String())}, knownhosts.Normalize(host)), key) + "\n")
		if e != nil {
			return e
		}
		fmt.Printf("Permanently added %#v to the list of known hosts.\n", host)
		return e

	}

	return fmt.Errorf("unknown error occured")
}

func CreateSession() error {
	c, e := dialPrivateKey()
	if strings.Contains(fmt.Sprintf("%v", e), "unable to authenticate") {
		c, e = dialPassword()
		if e != nil {
			return fmt.Errorf("unable to authenticate")
		}
	} else if e != nil {
		return fmt.Errorf("client creation error: %s", e)
	}
	s, e := c.NewSession()
	if e != nil {
		return fmt.Errorf("session create error: %s", e)
	}
	s.Stdout = os.Stdout
	s.Stderr = os.Stderr
	s.Stdin = os.Stdin
	m := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	f := int(os.Stdin.Fd())
	if terminal.IsTerminal(f) {
		o, e := terminal.MakeRaw(f)
		if e != nil {
			return e
		}
		defer terminal.Restore(f, o)
		w, h, e := terminal.GetSize(f)
		if e != nil {
			return e
		}
		e = s.RequestPty("xterm-256color", h, w, m)
		if e != nil {
			return e
		}
	}
	e = s.Shell()
	if e != nil {
		return e
	}
	s.Wait()
	return nil
}

func dialPrivateKey() (*ssh.Client, error) {
	s, e := prepareSigner(d.UserKey)
	if e != nil {
		return nil, fmt.Errorf("private key read error: %s", e)
	}

	config := &ssh.ClientConfig{
		User: d.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(s),
		},
		HostKeyCallback: verifySshServer,
	}
	c, e := ssh.Dial("tcp", d.Host+":"+d.Port, config)

	return c, e

}

func dialPassword() (*ssh.Client, error) {
	fmt.Printf("Password for %#v: ", d.User)
	bytePassword, e := term.ReadPassword(int(syscall.Stdin))
	if e != nil {
		return nil, e
	}
	userPass := string(bytePassword)
	config := &ssh.ClientConfig{
		User: d.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(userPass),
		},
		HostKeyCallback: verifySshServer,
	}
	c, e := ssh.Dial("tcp", d.Host+":"+d.Port, config)

	return c, e
}

func CheckArgs() error {
	userHome, _ := os.UserHomeDir()
	flag.StringVar(&d.UserKey, "i", userHome+"/"+defaultPrivateKeyFile, "User private key path")
	flag.StringVar(&d.HostKey, "k", userHome+"/"+defaultKnownHostsFile, "Known hosts path")
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
		if e := d.get(); e != nil {
			fmt.Println(e)
			return e
		}
		return nil
	}
}

func printUsage() {
	fmt.Printf("usage: tuffsh [-i identity_file] [-k known_hosts_file] [user@]destination[:port]\n")
}

func (d *destination) getPort() error {
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

func (d *destination) get() error {
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
