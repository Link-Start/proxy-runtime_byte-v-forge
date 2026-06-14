package lease

type Listener struct {
	ID       string
	Addr     string
	Protocol string
	Route    string
	Username string
	Password string
	Labels   map[string]string
}
