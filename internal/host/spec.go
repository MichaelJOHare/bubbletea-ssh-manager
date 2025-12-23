package host

// Spec is the shared representation of a host endpoint across the project.
//
// It maps directly to the subset of SSH-style config directives we support
// for both SSH and Telnet hosts.
type Spec struct {
	Alias    string // ssh-style Host alias from the config
	HostName string // hostname or IP address
	Port     string // port number as string
	User     string // user name
}
