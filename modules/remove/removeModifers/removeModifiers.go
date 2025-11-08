package removemodifers

import (
	"fmt"
	"strings"

	"dns-hostlist-compiler/modules/ruleUtils"
)

func RemoveModifiers(rules []string) []string {
	var filtered []string

	for _, rawRuleText := range rules {
		var ruleText string = strings.TrimSpace(rawRuleText)

		if len(ruleText) == 0 || ruleUtils.IsComment(ruleText) {
			filtered = append(filtered, ruleText)
			continue
		}

		props := ruleUtils.LoadAdblockRuleProperties(ruleText)
		if props.Pattern == "" {
			filtered = append(filtered, ruleText)
		}

		ruleUtils.RemoveModifier(&props, "third-party")
		ruleUtils.RemoveModifier(&props, "3p")
		ruleUtils.RemoveModifier(&props, "all")
		ruleUtils.RemoveModifier(&props, "document")
		ruleUtils.RemoveModifier(&props, "doc")
		ruleUtils.RemoveModifier(&props, "popup")
		filtered = append(filtered, ruleUtils.AdblockRuleToString(props))
	}

	fmt.Printf("removemodifiers - start: %d\tend: %d\n", len(rules), len(filtered))
	return filtered
}
