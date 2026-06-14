package sourceplane

import "time"

type Config struct {
	EgressProfiles      []EgressProfile
	Endpoint            Endpoint
	HealthCheckURL      string
	HealthCheckInterval time.Duration
	HealthCheckTimeout  time.Duration
}

type Endpoint struct {
	Addr     string
	Protocol string
}

type EgressProfile struct {
	ID          string
	DisplayName string
	Enabled     bool
	Line        EgressProfileLine
	Exit        EgressProfileExit
}

type EgressProfileLayer struct {
	Kind           string
	ProviderID     string
	ResourceID     string
	NodeID         string
	HealthCheckURL string
	HealthInterval time.Duration
	HealthTimeout  time.Duration
	ExpectedStatus uint32
}

type EgressProfileLine = EgressProfileLayer

type EgressProfileExit = EgressProfileLayer
