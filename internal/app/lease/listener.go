package lease

import (
	"errors"
	"strings"
)

const ListenerModeDynamicSessionLease = "dynamic_ip_session_lease"

var ErrListenerPasswordRequired = errors.New("dynamic lease listener password is not configured")

type Listener struct {
	ID       string
	Addr     string
	Protocol string
	Route    string
	Username string
	Password string
	Labels   map[string]string
}

type ListenerInput struct {
	ID        string
	Addr      string
	Protocol  string
	Route     string
	Username  string
	Password  string
	AccountID string
	LeaseID   string
}

func NewListener(input ListenerInput) (Listener, error) {
	if strings.TrimSpace(input.Password) == "" {
		return Listener{}, ErrListenerPasswordRequired
	}
	return Listener{
		ID:       input.ID,
		Addr:     input.Addr,
		Protocol: input.Protocol,
		Route:    input.Route,
		Username: input.Username,
		Password: input.Password,
		Labels: map[string]string{
			"mode":         ListenerModeDynamicSessionLease,
			LabelAccountID: input.AccountID,
			LabelLeaseID:   input.LeaseID,
		},
	}, nil
}
