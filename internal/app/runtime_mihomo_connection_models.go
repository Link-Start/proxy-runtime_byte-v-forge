package app

type mihomoConnectionsResponse struct {
	Connections []mihomoConnection `json:"connections"`
}

type mihomoConnection struct {
	ID          string                   `json:"id"`
	Rule        string                   `json:"rule"`
	RulePayload string                   `json:"rulePayload"`
	Chains      []string                 `json:"chains"`
	Metadata    mihomoConnectionMetadata `json:"metadata"`
}

type mihomoConnectionMetadata struct {
	InboundUser string `json:"inboundUser"`
}
