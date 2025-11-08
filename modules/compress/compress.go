package compress

import (
	"dns-hostlist-compiler/modules/ruleUtils"
	"fmt"
	"strings"
)

type BlocklistRule struct {
	RuleText         string
	CanCompress      bool
	Hostname         string
	OriginalRuleText string
}

func extractHostnames(hostname string) []string {
	var parts []string = strings.Split(hostname, ".")
	var domains []string
	for i := range parts {
		domains = append(domains, strings.Join(parts[i:], "."))
	}
	return domains
}

func toAdblockRules(ruleText string) []BlocklistRule {
	var adblockRules []BlocklistRule

	// /etc/hosts rules can be compressed
	if ruleUtils.IsEtcHostsRule(ruleText) {
		props, _ := ruleUtils.LoadEtcHostsRuleProperties(ruleText)

		for _, hostname := range props.Hostnames {
			adblockRules = append(adblockRules, BlocklistRule{
				RuleText:         fmt.Sprintf("||%s^", hostname),
				CanCompress:      true,
				Hostname:         hostname,
				OriginalRuleText: ruleText,
			})
		}

		return adblockRules
	}

	// simple domain names should also be compressed (and converted)
	if ruleUtils.IsJustDomain(ruleText) {
		return []BlocklistRule{{
			RuleText:         fmt.Sprintf("||%s^", ruleText),
			CanCompress:      true,
			Hostname:         ruleText,
			OriginalRuleText: ruleText,
		}}
	}

	props := ruleUtils.LoadAdblockRuleProperties(ruleText)
	if props.Hostname != "" && !props.Whitelist && len(props.Options) == 0 {
		adblockRules = append(adblockRules, BlocklistRule{
			RuleText:         ruleText,
			CanCompress:      true,
			Hostname:         props.Hostname,
			OriginalRuleText: ruleText,
		})

		return adblockRules
	}

	// Cannot parse or compress
	adblockRules = append(adblockRules, BlocklistRule{
		RuleText:         ruleText,
		CanCompress:      false,
		Hostname:         "",
		OriginalRuleText: ruleText,
	})

	return adblockRules
}

/**
 * This transformation compresses the final list by removing redundant rules.
 * Please note, that it also converts /etc/hosts rules into adblock-style rules.
 * 1. It converts all rules to adblock-style rules. For instance,
 * "0.0.0.0 example.org" will be converted to "||example.org^".
 * 2. It discards the rules that are already covered by existing rules.
 * For instance, "||example.org" blocks "example.org" and all it's subdomains,
 * therefore you don't need additional rules for the subdomains.
 */
func Compress(rules []string) []string {
	// var initialLength int = len(rules)

	var byHostname map[string]bool = make(map[string]bool)
	var filtered []BlocklistRule

	// First loop:
	// 1. Transform /etc/hosts rules to adblock-style rules
	// 2. Fill "byHostname" lookup table
	// 3. Check "byHostname" to eliminate duplicates on the first run
	for _, rule := range rules {
		var adblockRules []BlocklistRule = toAdblockRules(rule)
		for _, adblockRule := range adblockRules {
			if adblockRule.CanCompress {
				if _, exists := byHostname[adblockRule.Hostname]; !exists {
					filtered = append(filtered, adblockRule)
					byHostname[adblockRule.Hostname] = true
				}
			} else {
				filtered = append(filtered, adblockRule)
			}
		}
	}

	// Second loop:
	// 1. Extract all hostnames up to TLD+1
	// 2. Check them against "byHostname" and discard the rule
	// if it's already covered by an existing rule.
	for i := len(filtered) - 1; i >= 0; i -= 1 {
		var rule BlocklistRule = filtered[i]
		var discard bool = false

		if rule.CanCompress {
			var hostnames []string = extractHostnames(rule.Hostname)
			// Start iterating from 1 -- don't check the full hostname
			for j := 1; j < len(hostnames); j += 1 {
				var hostname string = hostnames[j]
				if byHostname[hostname] {
					discard = true
					break
				}
			}
		}

		if discard {
			filtered = append(filtered[:i], filtered[i+1:]...)
		}
	}

	compressedList := make([]string, len(filtered))
	for i, rule := range filtered {
		compressedList[i] = rule.RuleText
	}

	fmt.Printf("compress - start: %d\tend: %d\n", len(rules), len(compressedList))
	return compressedList
}
