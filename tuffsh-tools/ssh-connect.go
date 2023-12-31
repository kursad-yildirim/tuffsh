package tuffshtools

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"runtime"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
	"golang.org/x/term"
)

func InitiateSSH() error {
	if e := d.getPort(); e != nil {
		return e
	}
	if e := d.getUser(); e != nil {
		return e
	}
	if e := d.getHost(); e != nil {
		fmt.Println(e)
		return e
	}
	switch sshCommand {
	case "nothing":
		interactiveShell()
	default:
		runCommand()
	}
	return nil
}

func interactiveShell() error {
	c, e := createClient()
	if e != nil {
		return e
	}
	s, e := c.NewSession()
	if e != nil {
		return e
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
	if term.IsTerminal(f) {
		o, e := term.MakeRaw(f)
		if e != nil {
			return e
		}
		defer term.Restore(f, o)
		if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
			w, h, e = term.GetSize(f)
		}
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

func runCommand() error {
	c, e := createClient()
	if e != nil {
		return e
	}
	s, e := c.NewSession()
	if e != nil {
		return e
	}
	s.Stdout = os.Stdout
	s.Stderr = os.Stderr
	if e := s.Run(sshCommand); e != nil {
		return fmt.Errorf("failed to run: " + e.Error())
	}
	return nil
}

func createClient() (*ssh.Client, error) {
	c, e := dialPrivateKey()
	if strings.Contains(fmt.Sprintf("%v", e), "unable to authenticate") {
		c, e = dialPassword()
		if e != nil {
			return nil, fmt.Errorf("unable to authenticate")
		}
	} else if e != nil {
		return nil, fmt.Errorf("client creation error: %s", e)
	}
	return c, nil
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
