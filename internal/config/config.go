package config

type LoadBalanceStrategy string

const (
	RoundRobin LoadBalanceStrategy = "round_robin"
)

type Config struct {
	Server    ServerConfig `json:"server"`
	Upstreams []Upstream   `json:"upstreams"`
	Routes    []Route      `json:"routes"`
}

type ServerConfig struct {
	Listen int `json:"listen"`
}

type Upstream struct {
	Name     string              `json:"name"`
	Strategy LoadBalanceStrategy `json:"strategy"`
	Servers  []string            `json:"servers"`
}

type Route struct {
	Path     string `json:"path"`
	Upstream string `json:"upstream"`
}
