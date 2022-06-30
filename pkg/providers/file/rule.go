package file

import "sort"

// Rule holds the config for a rule in serialized form
type Rule struct {
	// Key is a string that must be unique to each rule, and must remain
	// the same when rules are updated.
	Key string `json:"key"`
	// Ports is the ports that are bound to this rule.
	Ports []int `json:"ports"`
	// Targets is the targets this app will proxy connections to.
	Targets []string `json:"targets"`
}

type RuleSlice []Rule

func (s RuleSlice) Len() int {
	return len(s)
}

func (s RuleSlice) Less(i, j int) bool {
	return s[i].Key < s[j].Key
}

func (s RuleSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// FindIndexByKey returns the index of such rule that have the corresponding key if it exists,
// otherwise return len(s). It is assumed that the RuleSlice are sorted already.
func (s RuleSlice) FindIndexByKey(key string) int {
	searchedIndex := sort.Search(len(s), func(i int) bool {
		return s[i].Key >= key
	})
	if searchedIndex == len(s) {
		return searchedIndex
	}
	if s[searchedIndex].Key == key {
		return searchedIndex
	} else {
		return len(s)
	}
}

func (s RuleSlice) ExistByKey(key string) bool {
	return s.FindIndexByKey(key) != len(s)
}
