package deduplicate

import (
	"dns-hostlist-compiler/modules/ruleUtils"
	"fmt"
)

func Deduplicate(rules []string) []string {
	if len(rules) == 0 {
		return rules
	}

	// Clone the original array before modifying it
	var filtered []string = rules
	var prevRuleRemoved bool = false
	var rulesIndex map[string]struct{} = make(map[string]struct{})

	for iFiltered := len(filtered) - 1; iFiltered >= 0; iFiltered -= 1 {
		var ruleText string = filtered[iFiltered]

		_, exists := rulesIndex[ruleText]
		if !exists {
			rulesIndex[ruleText] = struct{}{}
		}

		if exists && !ruleUtils.IsComment(ruleText) && len(ruleText) > 0 {
			prevRuleRemoved = true
			filtered = append(filtered[:iFiltered], filtered[iFiltered+1:]...)
		} else if prevRuleRemoved && (ruleUtils.IsComment(ruleText) || len(ruleText) == 0) {
			// Remove preceding comments and empty lines
			filtered = append(filtered[:iFiltered], filtered[iFiltered+1:]...)
		} else {
			// Stop removing comments
			prevRuleRemoved = false
		}
	}

	fmt.Printf("deduplicate - start: %d\tend: %d\n", len(rules), len(filtered))
	return filtered
}
