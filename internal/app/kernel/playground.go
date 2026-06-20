package kernel

// Well-known identity of the built-in playground egress profile, ingress rule
// and user, shared across settings normalization and lease/listener routing.
const (
	PlaygroundProfileID = "playground-egress"
	PlaygroundUsername  = "playground"
	PlaygroundRuleID    = "in-user-playground"
)
