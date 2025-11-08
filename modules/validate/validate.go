package validate

import (
	"dns-hostlist-compiler/modules/ruleUtils"
	"dns-hostlist-compiler/modules/utils"
	"fmt"
	"regexp"
	"strings"
)

var (
	HOSTNAME_REGEX      *regexp.Regexp      = regexp.MustCompile(`(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9][a-z0-9-]{0,61}[a-z0-9]`)
	SUPPORTED_MODIFIERS map[string]struct{} = map[string]struct{}{"important": {}, "~important": {}, "badfilter": {}, "ctag": {}, "denyallow": {}}
	MIN_PATTERN_LENGTH  int                 = 5
)

func validHostname(hostname, ruleText string) bool {
	return HOSTNAME_REGEX.MatchString(hostname)
	// Todo: Fix this to actually get the WHOIS data
	/*
		res, err := whoisLookup(hostname)
		if err != nil {
			return false
		}
		u, err := url.Parse("http://" + hostname)
		if err != nil {
			// hostname invalid
			return false
		}
		parsedHostname := u.Hostname()
		if err != nil || (parsedHostname == "") || (net.ParseIP(hostname) != nil) {
			// hostname invalid
			return false
		}
		publicSuffix := getPublicSuffix(parsedHostname)
		// matching whole public suffix not allowed
		return parsedHostname != publicSuffix
	*/
}

/*
func getPublicSuffix(hostname string) string {
	domainParts := strings.Split(hostname, ".")
	if len(domainParts) < 2 {
		return ""
	}
	return domainParts[len(domainParts)-2] + "." + domainParts[len(domainParts)-1]
}
*/

/**
 * Validates an /etc/hosts rule.
 *
 * We do one very simple thing:
 * 1. Validate all the hostnames
 * 2. Prohibit rules that block the whole public suffix
 * 3. Prohibit rules that contain invalid domain names
 */
func validEtcHostsRule(ruleText string) bool {
	props, err := ruleUtils.LoadEtcHostsRuleProperties(ruleText)
	if err != nil {
		return false
	}
	if len(props.Hostnames) == 0 {
		return false
	}

	for _, hostname := range props.Hostnames {
		if !validHostname(hostname, ruleText) {
			return false
		}
	}

	return true
}

/**
 * Validates an adblock-style rule.
 *
 * 1. It checks if the rule contains only supported modifiers.
 * 2. It checks whether the pattern is not too wide (should be at least 5 characters).
 * 3. If checks if the pattern does not contain characters that cannot be in a domain name.
 * 4. For domain-blocking rules like ||domain^ it checks that the domain is
 * valid and does not block too much.
 */
func validAdblockRule(ruleText string) bool {
	props := ruleUtils.LoadAdblockRuleProperties(ruleText)
	if props.Pattern == "" {
		return false
	}

	// 1. It checks if the rule contains only supported modifiers.
	if len(props.Options) > 0 {
		for _, option := range props.Options {
			if _, exists := SUPPORTED_MODIFIERS[option.Name]; !exists {
				return false
			}
		}
	}

	// 2. It checks whether the pattern is not too wide (should be at least 5 characters).
	if len(props.Pattern) < MIN_PATTERN_LENGTH {
		return false
	}

	// 3. If checks if the pattern does not contain characters that cannot be in a domain name.
	// 3.1. Special case: regex rules
	// Do nothing with regex rules -- they may contain all kinds of special chars
	if strings.HasPrefix(props.Pattern, "/") && strings.HasSuffix(props.Pattern, "/") {
		return true
	}

	// However, regular adblock-style rules if they match a domain name
	// a-zA-Z0-9- -- permitted in the domain name
	// *|^ -- special characters used by adblock-style rules
	// One more special case is rules starting with ://s
	toTest := props.Pattern
	toTest = strings.TrimPrefix(toTest, "://")

	var checkChars bool = regexp.MustCompile(`^[a-zA-Z0-9-.*|^]+$`).Match([]byte(toTest))
	if !checkChars {
		return false
	}

	// 4. Validate domain name
	// Note that we don't check rules that contain wildcard characters
	var sepIdx int = strings.Index(props.Pattern, "^")
	var wildcardIdx int = strings.Index(props.Pattern, "*")
	if sepIdx != -1 && wildcardIdx != -1 && wildcardIdx > sepIdx {
		// Smth like ||example.org^test* -- invalid
		return false
	}

	if strings.HasPrefix(props.Pattern, "||") && sepIdx != -1 && wildcardIdx != -1 {
		var hostname string = utils.SubstringBetween(ruleText, "||", "^")
		if !validHostname(hostname, ruleText) {
			return false
		}

		// If there's something after ^ in the pattern - something went wrong
		// unless it's `^|` which is a rather often case
		if (len(props.Pattern) > (sepIdx + 1)) && props.Pattern[sepIdx+1] != '|' {
			return false
		}
	}

	return true
}

/**
 * Validates the rule.
 *
 * Emptry strings and comments are considered valid.
 *
 * For /etc/hosts rules: utils.validEtcHostsRule
 * For adblock-style rules: utils.validAdblockRule
 */
func valid(ruleText string) bool {
	if ruleUtils.IsComment(ruleText) || len(strings.TrimSpace(ruleText)) == 0 {
		return true
	}

	if ruleUtils.IsEtcHostsRule(ruleText) {
		return validEtcHostsRule(ruleText)
	}

	return validAdblockRule(ruleText)
}

/**
 * Validates all rules
 */
func Validate(rules []string) []string {
	var filtered []string = rules
	var prevRuleRemoved bool = false

	for iFiltered := len(filtered) - 1; iFiltered >= 0; iFiltered -= 1 {
		var ruleText string = filtered[iFiltered]

		if !valid(ruleText) {
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

	fmt.Printf("validate - start: %d\tend: %d\n", len(rules), len(filtered))
	return filtered
}
