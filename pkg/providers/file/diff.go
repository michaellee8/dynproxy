package file

import (
	"github.com/michaellee8/dynproxy/pkg/proxy/op"
	"github.com/michaellee8/dynproxy/pkg/utils"
	"sort"
)

func DiffConfig(oldCfg Config, newCfg Config) (ops []op.Operation) {
	sort.Sort(RuleSlice(oldCfg.Rules))
	sort.Sort(RuleSlice(newCfg.Rules))

	for i := range oldCfg.Rules {
		sort.Ints(oldCfg.Rules[i].Ports)
		sort.Strings(oldCfg.Rules[i].Targets)
	}

	for i := range newCfg.Rules {
		sort.Ints(newCfg.Rules[i].Ports)
		sort.Strings(newCfg.Rules[i].Targets)
	}

	commonRulesFromOldCfg := RuleSlice{}
	commonRulesFromNewCfg := RuleSlice{}

	for _, rule := range oldCfg.Rules {
		if RuleSlice(newCfg.Rules).ExistByKey(rule.Key) {
			commonRulesFromOldCfg = append(commonRulesFromOldCfg, rule)
		} else {
			ops = append(ops, op.RemoveRuleOperation(rule.Key))
		}
	}

	for _, rule := range newCfg.Rules {
		if RuleSlice(oldCfg.Rules).ExistByKey(rule.Key) {
			commonRulesFromNewCfg = append(commonRulesFromNewCfg, rule)
		} else {
			ops = append(ops, op.AddRuleOperation(rule.Key))
			for _, target := range rule.Targets {
				ops = append(ops, op.AddTargetOperation(rule.Key, target))
			}
			for _, port := range rule.Ports {
				ops = append(ops, op.AddPortOperation(rule.Key, port))
			}
		}
	}

	if len(commonRulesFromOldCfg) != len(commonRulesFromNewCfg) {
		panic("assert failure: common rules derived by different methods should be same len")
	}

	for ruleIndex := range commonRulesFromOldCfg {
		oldRule := commonRulesFromOldCfg[ruleIndex]
		newRule := commonRulesFromNewCfg[ruleIndex]

		if oldRule.Key != newRule.Key {
			panic("assert failure: old and new rule must have same key")
		}

		ruleKey := oldRule.Key

		for _, port := range oldRule.Ports {
			if !utils.HasIntOnSorted(newRule.Ports, port) {
				ops = append(ops, op.RemovePortOperation(ruleKey, port))
			}
		}

		for _, port := range newRule.Ports {
			if !utils.HasIntOnSorted(oldRule.Ports, port) {
				ops = append(ops, op.AddPortOperation(ruleKey, port))
			}
		}

		for _, target := range oldRule.Targets {
			if !utils.HasStringOnSorted(newRule.Targets, target) {
				ops = append(ops, op.RemoveTargetOperation(ruleKey, target))
			}
		}

		for _, target := range newRule.Targets {
			if !utils.HasStringOnSorted(oldRule.Targets, target) {
				ops = append(ops, op.AddTargetOperation(ruleKey, target))
			}
		}
	}

	sort.Stable(op.OperationSlice(ops))

	return
}
