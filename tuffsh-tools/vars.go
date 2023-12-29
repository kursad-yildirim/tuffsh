package tuffshtools

const defaultPrivateKeyFile = ".ssh/id_rsa"
const defaultKnownHostsFile = ".ssh/known_hosts"
const defaultSshPort = "22"

type destination struct {
	User    string
	Host    string
	Port    string
	UserKey string
	HostKey string
}

var help bool
var d destination
