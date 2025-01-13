package config

type ConfigRaft struct {
	NodeID   string `json:"nodeID"`
	BindAddr string `json:"bindAddr"`
	DataDir  string `json:"dataDir"`
}

type ConfigAPI struct {
	Port int `json:"port"`
}

type ConfigMembership struct {
	NodeName       string            `json:"nodeName"`
	BindAddr       string            `json:"bindAddr"`
	Tags           map[string]string `json:"tags"`
	StartJoinAddrs []string          `json:"startJoinAddrs"`
}
