package removecomments

import (
	"dns-hostlist-compiler/modules/ruleUtils"
	"fmt"
)

func RemoveComments(rules []string) []string {
	var filtered []string
	for _, rule := range rules {
		if !ruleUtils.IsComment(rule) {
			filtered = append(filtered, rule)
		}
	}

	fmt.Printf("removecomments - start: %d\tend: %d\n", len(rules), len(filtered))
	return filtered
}
