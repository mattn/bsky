package main

import (
	"regexp"
	"strings"
)

const (
	urlPattern     = `https?://[-A-Za-z0-9+&@#\/%?=~_|!:,.;\(\)]+`
	mentionPattern = `@[a-zA-Z0-9.]+`
	tagPattern     = `\B#\S+`
)

var (
	urlRe     = regexp.MustCompile(urlPattern)
	mentionRe = regexp.MustCompile(mentionPattern)
	tagRe     = regexp.MustCompile(tagPattern)
)

type entry struct {
	start int64
	end   int64
	text  string
}

func extractLinks(text string) []entry {
	var result []entry
	matches := urlRe.FindAllStringSubmatchIndex(text, -1)
	for _, m := range matches {
		result = append(result, entry{
			text:  text[m[0]:m[1]],
			start: int64(len([]rune(text[0:m[0]]))),
			end:   int64(len([]rune(text[0:m[1]])))},
		)
	}
	return result
}

func extractLinksBytes(text string) []entry {
	var result []entry
	matches := urlRe.FindAllStringSubmatchIndex(text, -1)
	for _, m := range matches {
		result = append(result, entry{
			text:  text[m[0]:m[1]],
			start: int64(len(text[0:m[0]])),
			end:   int64(len(text[0:m[1]]))},
		)
	}
	return result
}

func extractMentions(text string) []entry {
	var result []entry
	matches := mentionRe.FindAllStringSubmatchIndex(text, -1)
	for _, m := range matches {
		result = append(result, entry{
			text:  strings.TrimPrefix(text[m[0]:m[1]], "@"),
			start: int64(len([]rune(text[0:m[0]]))),
			end:   int64(len([]rune(text[0:m[1]])))},
		)
	}
	return result
}

func extractMentionsBytes(text string) []entry {
	var result []entry
	matches := mentionRe.FindAllStringSubmatchIndex(text, -1)
	for _, m := range matches {
		result = append(result, entry{
			text:  strings.TrimPrefix(text[m[0]:m[1]], "@"),
			start: int64(len(text[0:m[0]])),
			end:   int64(len(text[0:m[1]]))},
		)
	}
	return result
}

func extractTags(text string) []entry {
	var result []entry
	matches := tagRe.FindAllStringSubmatchIndex(text, -1)
	for _, m := range matches {
		result = append(result, entry{
			text:  strings.TrimPrefix(text[m[0]:m[1]], "#"),
			start: int64(len([]rune(text[0:m[0]]))),
			end:   int64(len([]rune(text[0:m[1]])))},
		)
	}
	return result
}

func extractTagsBytes(text string) []entry {
	var result []entry
	matches := tagRe.FindAllStringSubmatchIndex(text, -1)
	for _, m := range matches {
		result = append(result, entry{
			text:  strings.TrimPrefix(text[m[0]:m[1]], "#"),
			start: int64(len(text[0:m[0]])),
			end:   int64(len(text[0:m[1]]))},
		)
	}
	return result
}
