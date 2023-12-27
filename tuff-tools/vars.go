package tuff

const defaultPrivateKeyFile = "./.ssh/id_rsa"
const defaultKnownHostsFile = "./.ssh/known_hosts"

type Destination struct {
	User    string
	Host    string
	Port    string
	UserKey string
	HostKey string
}

var help bool
var D Destination
