package tuffshtools

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

func prepareSigner() (ssh.Signer, error) {
	k, e := os.ReadFile(defaultPrivateKeyFile)
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
	callback, e := knownhosts.New(defaultKnownHostsFile)
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
		f, e := os.OpenFile(defaultKnownHostsFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
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

func createSession(h, p string) (*ssh.Session, error) {
	signer, e := prepareSigner()
	if e != nil {
		return nil, fmt.Errorf("private key read error: %s", e)
	}
	config := &ssh.ClientConfig{
		User: "trip",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: verifySshServer,
	}
	client, e := ssh.Dial("tcp", h+":"+p, config)
	if e != nil {
		return nil, fmt.Errorf("client creation error: %s", e)
	}
	session, e := client.NewSession()
	if e != nil {
		return nil, fmt.Errorf("session create error: %s", e)
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	if e := session.RequestPty("vt100", 80, 40, modes); e != nil {
		return nil, fmt.Errorf("pty error: %s", e)
	}

	return session, nil
}

func TuffSSH(h, p string) (chan<- string, <-chan string, error) {
	in := make(chan string, 1)
	out := make(chan string, 1)
	var wg sync.WaitGroup

	session, e := createSession(h, p)
	if e != nil {
		return nil, nil, fmt.Errorf("session create  failed: %s", e)
	}
	w, e := session.StdinPipe()
	if e != nil {
		return nil, nil, fmt.Errorf("StdinPipe error: %s", e)
	}
	r, e := session.StdoutPipe()
	if e != nil {
		return nil, nil, fmt.Errorf("StdoutPipe error: %s", e)
	}
	wg.Add(1)
	go func() {
		for cmd := range in {
			wg.Add(1)
			w.Write([]byte(cmd + "\n"))
			wg.Wait()
		}
	}()
	go func() {
		var buffer [16 * 1024]byte
		for {
			n, err := r.Read(buffer[:])
			if err != nil {
				close(in)
				close(out)
				return
			}
			if string(buffer[n-2:n]) == "\r\n" {
				out <- string(buffer[:n])
			}
			if buffer[n-2] == '$' || buffer[n-2] == '#' {
				out <- string(buffer[:n])
				wg.Done()
			}
		}
	}()

	if err := session.Start("/bin/bash"); err != nil {
		return nil, nil, fmt.Errorf("session start failed %#v", err)
	}
	return in, out, nil

}
