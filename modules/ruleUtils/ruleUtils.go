package ruleUtils

import (
	"dns-hostlist-compiler/modules/utils"
	"fmt"
	"log"
	"regexp"
	"strings"
)

type AdblockRuleTokens struct {
	Pattern   string
	Options   string
	Whitelist bool
}

type EtcHostsRule struct {
	RuleText  string
	Hostnames []string
}

type ruleOption struct {
	Name  string
	Value string
}

type AdblockRule struct {
	RuleText  string
	Pattern   string
	Options   []ruleOption
	Whitelist bool
	Hostname  string
}

var (
	// domainRegex *regexp.Regexp = regexp.MustCompilePOSIX(`^(?=.{1,255}$)[0-9A-Za-z](?:(?:[0-9A-Za-z]|-){0,61}[0-9A-Za-z])?(?:\.[0-9A-Za-z](?:(?:[0-9A-Za-z]|-){0,61}[0-9A-Za-z])?)*\.?$`)
	// Perl equivalent:
	domainRegex   *regexp.Regexp = regexp.MustCompile(`^([0-9A-Za-z](?:[0-9A-Za-z-]{0,61}[0-9A-Za-z])?)(\.[0-9A-Za-z](?:[0-9A-Za-z-]{0,61}[0-9A-Za-z])?)*$`)
	etcHostsRegex *regexp.Regexp = regexp.MustCompile(`^([a-f0-9.:\][]+)(%[a-z0-9]+)?\s+([^#]+)(#.*)?$`)
)

func IsComment(ruleText string) bool {
	return (strings.TrimSpace(ruleText) == "" || strings.HasPrefix(ruleText, "#") || strings.HasPrefix(ruleText, "!") || strings.HasPrefix(ruleText, "####"))
}

func IsAllowRule(ruleText string) bool {
	return strings.HasPrefix(ruleText, "@@")
}

func IsJustDomain(ruleText string) bool {
	return strings.Contains(ruleText, ".") && domainRegex.Match([]byte(ruleText))
}

func IsEtcHostsRule(ruleText string) bool {
	return etcHostsRegex.Match([]byte(ruleText))
}

func parseRuleTokens(ruleText string) AdblockRuleTokens {
	var tokens AdblockRuleTokens = AdblockRuleTokens{
		Pattern:   "",
		Options:   "",
		Whitelist: false,
	}
	var startIndex int = 0

	if IsAllowRule(ruleText) {
		tokens.Whitelist = true
		startIndex = 2
	}

	if len(ruleText) <= startIndex {
		log.Fatalf("parseRuleTokens - the rule is too short: %s", ruleText)
	}

	// Setting pattern to rule text (for the case of empty options)
	tokens.Pattern = ruleText[startIndex:]

	// Avoid parsing options inside of a regex rule
	if strings.HasPrefix(tokens.Pattern, "/") && strings.HasSuffix(tokens.Options, "/") && !strings.Contains(tokens.Options, "replace=") {
		return tokens
	}

	for i := len(ruleText) - 1; i >= startIndex; i -= 1 {
		var c byte = ruleText[i]
		if c == '$' {
			if i > startIndex && ruleText[i-1] == '\\' {
				// Escaped, do nothing
			} else {
				tokens.Pattern = ruleText[startIndex:i]
				tokens.Options = ruleText[i+1:]
				break
			}
		}
	}

	return tokens
}

func LoadEtcHostsRuleProperties(ruleText string) (EtcHostsRule, error) {
	var rule string = strings.TrimSpace(ruleText)
	if strings.Contains(rule, "#") {
		rule = rule[0:strings.Index(rule, "#")]
	}

	var hostnames []string = strings.Fields(rule)
	hostnames = hostnames[1:]
	if len(hostnames) < 1 {
		return EtcHostsRule{}, fmt.Errorf("LoadEtcHostsRuleProperties - invalid /etc/hosts rule: %s", ruleText)
	}

	return EtcHostsRule{RuleText: ruleText, Hostnames: hostnames}, nil
}

func extractHostname(pattern string) string {
	var hostnameRegex *regexp.Regexp = regexp.MustCompile(`^\|\|([a-z0-9-.]+)\^$`)
	var matches []string = hostnameRegex.FindStringSubmatch(pattern)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func LoadAdblockRuleProperties(ruleText string) AdblockRule {
	var tokens AdblockRuleTokens = parseRuleTokens(strings.TrimSpace(ruleText))
	var rule AdblockRule = AdblockRule{
		RuleText:  ruleText,
		Pattern:   tokens.Pattern,
		Whitelist: tokens.Whitelist,
		Options:   []ruleOption{},
		Hostname:  extractHostname(tokens.Pattern),
	}

	if len(tokens.Options) > 0 {
		optionParts := utils.SplitByDelimiterWithEscapeCharacter(tokens.Options, ',', '\\', false)
		if len(optionParts) > 0 {
			for _, option := range optionParts {
				var parts []string = strings.SplitN(option, "=", 2)
				var name string = parts[0]
				var value string
				if len(parts) > 1 {
					value = parts[1]
				} else {
					value = ""
				}
				rule.Options = append(rule.Options, ruleOption{Name: name, Value: value})
			}
		}
	}

	return rule
}

func FindModifier(ruleProps AdblockRule, name string) *ruleOption {
	if ruleProps.Options == nil {
		return nil
	}

	for i := range ruleProps.Options {
		if ruleProps.Options[i].Name == name {
			return &ruleProps.Options[i]
		}
	}

	return nil
}

func RemoveModifier(ruleProps *AdblockRule, name string) bool {
	if ruleProps == nil || ruleProps.Options == nil {
		return false
	}

	var found bool = false
	for iOptions := len(ruleProps.Options) - 1; iOptions >= 0; iOptions -= 1 {
		if ruleProps.Options[iOptions].Name == name {
			ruleProps.Options = append(ruleProps.Options[:iOptions], ruleProps.Options[iOptions+1:]...)
			found = true
		}
	}

	return found
}

func AdblockRuleToString(ruleProps AdblockRule) string {
	var ruleText string = ""
	if ruleProps.Whitelist {
		ruleText = "@@"
	}
	ruleText += ruleProps.Pattern

	if len(ruleProps.Options) > 0 {
		ruleText += "$"
		for i, option := range ruleProps.Options {
			ruleText += option.Name
			if len(option.Value) > 0 {
				ruleText += "="
				ruleText += option.Value
			}
			if i < len(ruleProps.Options)-1 {
				ruleText += ","
			}
		}
	}

	return ruleText
}
