package rule

type Rule struct {
	// Key is a string that must be unique to each rule, and must remain
	// the same when rules are updated.
	Key string `json:"key"`
	// Ports is the ports that are bound to this rule.
	Ports []int `json:"ports"`
	// Targets is the targets this app will proxy connections to.
	Targets []string `json:"targets"`
}
