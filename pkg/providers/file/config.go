package file

type Config struct {
	// Rules are the rules of the proxy
	Rules []Rule `json:"rules"`
}
