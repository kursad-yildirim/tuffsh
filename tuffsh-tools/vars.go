package tuffshtools

const defaultPrivateKeyFile = ".ssh/id_rsa"
const defaultKnownHostsFile = ".ssh/known_hosts"
const defaultSshPort = "22"
const versionStr = "tuffsh 1.0"

var help, version bool

type destination struct {
	User    string
	Host    string
	Port    string
	UserKey string
	HostKey string
}

var d destination
var sshCommand string = "nothing"
var w, h int = 80, 40 // terminal size variables
