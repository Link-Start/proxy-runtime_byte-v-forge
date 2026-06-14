package mihomo

type mihomoListener struct {
	Name   string       `json:"name"`
	Type   string       `json:"type"`
	Listen string       `json:"listen,omitempty"`
	Port   int          `json:"port"`
	Rule   string       `json:"rule,omitempty"`
	Proxy  string       `json:"proxy,omitempty"`
	UDP    bool         `json:"udp"`
	Users  []mihomoUser `json:"users,omitempty"`
}

type mihomoUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
