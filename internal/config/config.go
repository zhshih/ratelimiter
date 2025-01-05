package config

type ConfigRaft struct {
	NodeID   string `json:"nodeID"`
	BindAddr string `json:"bindAddr"`
	DataDir  string `json:"dataDir"`
}

type ConfigAPI struct {
	Port int `json:"port"`
}
