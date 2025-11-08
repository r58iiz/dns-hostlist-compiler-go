package pipeline

import (
	"dns-hostlist-compiler/modules/compress"
	"dns-hostlist-compiler/modules/deduplicate"
	removecomments "dns-hostlist-compiler/modules/remove/removeComments"
	removemodifers "dns-hostlist-compiler/modules/remove/removeModifers"
	"dns-hostlist-compiler/modules/utils"
	"dns-hostlist-compiler/modules/validate"
	"fmt"
	"regexp"
)

func DedupeSlice[T comparable](sliceList []T) []T {
	dedupeMap := make(map[T]struct{})
	list := []T{}
	for _, slice := range sliceList {
		if _, exists := dedupeMap[slice]; !exists {
			dedupeMap[slice] = struct{}{}
			list = append(list, slice)
		}
	}
	return list
}

func RunPipeline(links []string) ([]string, error) {
	var rules []string
	re := regexp.MustCompile(`\r?\n`)

	for _, l := range links {
		res, err := utils.Download(l)
		if err != nil {
			return nil, fmt.Errorf("unable to download %s: %w", l, err)
		}
		parts := re.Split(res, -1)
		rules = append(rules, parts...)
	}

	// Process pipeline
	rules = removecomments.RemoveComments(rules)
	rules = compress.Compress(rules)
	rules = removemodifers.RemoveModifiers(rules)
	rules = validate.Validate(rules)
	rules = deduplicate.Deduplicate(rules)

	return rules, nil
}
