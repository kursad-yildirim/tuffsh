package tuffshtools

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"syscall"
	"tuffsh/tuff-tools"

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
	callback, e := knownhosts.New(tuff.D.HostKey)
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
		f, e := os.OpenFile(tuff.D.HostKey, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
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

func CreateSession(d tuff.Destination) error {
	signer, e := prepareSigner(d.UserKey)
	if e != nil {
		return fmt.Errorf("private key read error: %s", e)
	}
	config := &ssh.ClientConfig{
		User: d.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: verifySshServer,
	}
	client, e := ssh.Dial("tcp", d.Host+":"+d.Port, config)
	if strings.Contains(fmt.Sprintf("%v", e), "unable to authenticate") {
		fmt.Printf("Password for %#v: ", d.User)
		bytePassword, e := term.ReadPassword(int(syscall.Stdin))
		if e != nil {
			return e
		}
		userPass := string(bytePassword)
		config := &ssh.ClientConfig{
			User: d.User,
			Auth: []ssh.AuthMethod{
				ssh.Password(userPass),
			},
			HostKeyCallback: verifySshServer,
		}
		client, e = ssh.Dial("tcp", d.Host+":"+d.Port, config)
		if e != nil {
			return fmt.Errorf("client creation error: %s", e)
		}
	} else if e != nil {
		return fmt.Errorf("client creation error: %s", e)
	}
	session, e := client.NewSession()
	if e != nil {
		return fmt.Errorf("session create error: %s", e)
	}
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	fileDescriptor := int(os.Stdin.Fd())
	if terminal.IsTerminal(fileDescriptor) {
		originalState, e := terminal.MakeRaw(fileDescriptor)
		if e != nil {
			return e
		}
		defer terminal.Restore(fileDescriptor, originalState)

		termWidth, termHeight, e := terminal.GetSize(fileDescriptor)
		if e != nil {
			return e
		}

		e = session.RequestPty("xterm-256color", termHeight, termWidth, modes)
		if e != nil {
			return e
		}
	}

	e = session.Shell()
	if e != nil {
		return e
	}
	session.Wait()
	return nil
}
