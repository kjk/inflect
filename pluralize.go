package inflect

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

type rxRule struct {
	// TODO: for debugging, maybe remove when working
	rxStrJs string
	rxStrGo string

	rx          *regexp.Regexp
	replacement string
}

// Rule storage - pluralize and singularize need to be run sequentially,
// while other rules can be optimized using an object for instant lookups.
var pluralRules []rxRule
var singularRules []rxRule
var irregularPlurals = map[string]string{}
var irregularSingles = map[string]string{}
var uncountables = map[string]string{}

func init() {
	// order is important
	addIrregularRules()
	addPluralizationRules()
	addSingularizationRules()
	addUncountableRules()
}

// Add a pluralization rule to the collection.
func addPluralRule(rule string, replacement string) {
	rx, rxStrGo := sanitizeRule(rule)
	r := rxRule{
		rxStrJs:     rule,
		rxStrGo:     rxStrGo,
		rx:          rx,
		replacement: jsReplaceSyntaxToGo(replacement),
	}
	pluralRules = append(pluralRules, r)
}

func panicIf(cond bool, format string, args ...interface{}) {
	if !cond {
		return
	}
	s := format
	if len(args) > 0 {
		s = fmt.Sprintf(format, args...)
	}
	panic(s)
}

var (
	unicodeSyntaxRx = regexp.MustCompile(`\\u([[:xdigit:]]{4})`)
)

// best-effort of converting javascript regex syntax to equivalent go syntax
func jsRxSyntaxToGo(rx string) string {
	s := rx
	caseInsensitive := false
	panicIf(s[0] != '/', "expected '%s' to start with '/'", rx)
	s = s[1:]
	n := len(s)
	if s[n-1] == 'i' {
		n--
		caseInsensitive = true
		s = s[:n]
	}
	panicIf(s[n-1] != '/', "expected '%s' to end with '/'", rx)
	s = s[:n-1]
	// \uNNNN syntax for unicode code points to \x{NNNN} syntax for hex character code
	s = unicodeSyntaxRx.ReplaceAllString(s, "\\x{$1}")
	if caseInsensitive {
		s = "(?i)" + s
	}
	return s
}

func jsReplaceSyntaxToGo(s string) string {
	s = strings.Replace(s, "$0", "${0}", -1)
	s = strings.Replace(s, "$1", "${1}", -1)
	s = strings.Replace(s, "$2", "${2}", -1)
	return s
}

// Sanitize a pluralization rule to a usable regular expression.
func sanitizeRule(rule string) (*regexp.Regexp, string) {
	// in JavaScript, regexpes start with /
	// others are just regular strings
	var s string
	if rule[0] != '/' {
		// a plain string match is converted to regexp that:
		// ^ ... $ : does exact match (matches at the beginning and end)
		// (?i) : is case-insensitive
		s = `(?i)^` + rule + `$`
	} else {
		s = jsRxSyntaxToGo(rule)
	}
	return regexp.MustCompile(s), s
}

// Add a singularization rule to the collection.
func addSingularRule(rule, replacement string) {
	rx, rxGo := sanitizeRule(rule)
	r := rxRule{
		rxStrJs:     rule,
		rxStrGo:     rxGo,
		rx:          rx,
		replacement: jsReplaceSyntaxToGo(replacement),
	}
	singularRules = append(singularRules, r)
}

// copied from strings.ToUpper
// returns true if s is uppercase
func isUpper(s string) bool {
	isASCII, hasLower := true, false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= utf8.RuneSelf {
			isASCII = false
			break
		}
		hasLower = hasLower || (c >= 'a' && c <= 'z')
	}
	if isASCII {
		return !hasLower
	}
	for r := range s {
		if !unicode.IsUpper(rune(r)) {
			return false
		}
	}
	return true
}

// Pass in a word token to produce a function that can replicate the case on
// another word.
func restoreCase(word string, token string) string {
	// Tokens are an exact match.
	if word == token {
		return token
	}

	// Upper cased words. E.g. "HELLO".
	if isUpper(word) {
		return strings.ToUpper(token)
	}

	// Title cased words. E.g. "Title".
	prefix := word[:1]
	if isUpper(prefix) {
		return strings.ToUpper(token[:1]) + strings.ToLower(token[1:])
	}

	// Lower cased words. E.g. "test".
	return strings.ToLower(token)
}

// Replace a word using a rule.
func replace(word string, rule rxRule) string {
	// TODO: not sure if this covers all possibilities
	repl := rule.replacement
	if isUpper(word) {
		repl = strings.ToUpper(repl)
	}
	return rule.rx.ReplaceAllString(word, repl)
}

// Sanitize a word by passing in the word and sanitization rules.
func sanitizeWord(token string, word string, rules []rxRule) string {
	// Empty string or doesn't need fixing.
	if len(token) == 0 {
		return word
	}
	if _, ok := uncountables[token]; ok {
		return word
	}

	// Iterate over the sanitization rules and use the first one to match.
	// important that we iterate from the end
	n := len(rules)
	for i := n - 1; i >= 0; i-- {
		rule := rules[i]
		if rule.rx.MatchString(word) {
			return replace(word, rule)
		}
	}
	return word
}

// Replace a word with the updated word.
func replaceWord(word string, replaceMap map[string]string, keepMap map[string]string, rules []rxRule) string {
	// Get the correct token and case restoration functions.
	token := strings.ToLower(word)

	// Check against the keep object map.
	if _, ok := keepMap[token]; ok {
		return restoreCase(word, token)
	}

	// Check against the replacement map for a direct word replacement.
	if s, ok := replaceMap[token]; ok {
		return restoreCase(word, s)
	}

	// Run all the rules against the word.
	return sanitizeWord(token, word, rules)
}

// Check if a word is part of the map.
func checkWord(word string, replaceMap map[string]string, keepMap map[string]string, rules []rxRule) bool {
	token := strings.ToLower(word)

	if _, ok := keepMap[token]; ok {
		return true
	}

	if _, ok := replaceMap[token]; ok {
		return false
	}

	return sanitizeWord(token, token, rules) == token
}

// Add an irregular word definition.
func addIrregularRules() {
	for _, rule := range irregularRules {
		single := strings.ToLower(rule[0])
		plural := strings.ToLower(rule[1])

		irregularSingles[single] = plural
		irregularPlurals[plural] = single
	}
}

func addSingularizationRules() {
	for _, r := range singularizationRules {
		addSingularRule(r[0], r[1])
	}
}

func addUncountableRules() {
	for _, word := range uncountableRules {
		if word[0] != '/' {
			word = strings.ToLower(word)
			uncountables[word] = word
			continue
		}
		// Set singular and plural references for the word.
		addPluralRule(word, "$0")
		addSingularRule(word, "$0")
	}
}

func addPluralizationRules() {
	for _, rule := range pluralizationRules {
		addPluralRule(rule[0], rule[1])
	}
}

// Pluralize or singularize a word based on the passed in count.
func Pluralize(word string, count int, inclusive bool) string {
	var res string
	if count == 1 {
		res = ToSingular(word)
	} else {
		res = ToPlural(word)
	}

	if inclusive {
		return strconv.Itoa(count) + " " + res
	}
	return res
}

// IsPlural retruns true if word is plural
func IsPlural(word string) bool {
	return checkWord(word, irregularSingles, irregularPlurals, pluralRules)
}

// ToSingular singularizes a word.
func ToSingular(word string) string {
	return replaceWord(word, irregularPlurals, irregularSingles, singularRules)
}

// IsSingular returns true if a word is singular
func IsSingular(word string) bool {
	return checkWord(word, irregularPlurals, irregularSingles, singularRules)
}

// ToPlural makes a pluralized version of a word
func ToPlural(word string) string {
	return replaceWord(word, irregularSingles, irregularPlurals, pluralRules)
}
