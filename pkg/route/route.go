package route

import (
	"regexp"
	"strings"
	"sync"
)

var (
	regexen   = make(map[string]*regexp.Regexp)
	variables = regexp.MustCompile(`(\{[a-z][a-zA-Z0-9]+?\}|\*)`)

	methods = []string{
		"ACL",
		"BASELINE-CONTROL",
		"BIND",
		"CHECKIN",
		"CHECKOUT",
		"CONNECT",
		"COPY",
		"DELETE",
		"GET",
		"HEAD",
		"LABEL",
		"LINK",
		"LOCK",
		"MERGE",
		"MKACTIVITY",
		"MKCALENDAR",
		"MKCOL",
		"MKREDIRECTREF",
		"MKWORKSPACE",
		"MOVE",
		"OPTIONS",
		"ORDERPATCH",
		"PATCH",
		"POST",
		"PRI",
		"PROPFIND",
		"PROPPATCH",
		"PUT",
		"REBIND",
		"REPORT",
		"SEARCH",
		"TRACE",
		"UNBIND",
		"UNCHECKOUT",
		"UNLINK",
		"UNLOCK",
		"UPDATE",
		"UPDATEREDIRECTREF",
		"VERSION-CONTROL",
	}

	relock sync.Mutex
)

// Match will match given route towards given pattern
func Match(pattern, method, path string) (bool, error) {
	var fm string
	for _, m := range methods {
		met := strings.ToLower(m)
		if strings.HasPrefix(strings.ToLower(pattern), met+":") {
			fm = met
			pattern = pattern[len(m)+1:]
			break
		}
	}
	if len(fm) > 0 && len(method) > 0 && !strings.EqualFold(fm, method) {
		return false, nil
	}
	var parts []string
	var last int
	for _, x := range variables.FindAllStringIndex(pattern, -1) {
		if len(x) != 2 {
			continue
		}
		i, j := x[0], x[1]
		parts = append(parts, regexp.QuoteMeta(pattern[last:i]))
		switch {
		case pattern[i] == '{' && pattern[j-1] == '}':
			parts = append(parts, "([^/]+)")
		case pattern[i] == '*':
			parts = append(parts, "(.*)")
		default:
			parts = append(parts, pattern[i:j])
		}
		last = j
	}
	pettern := strings.Join(append(parts, pattern[last:]), "")
	reg, err := compileCached(pettern)
	if err != nil {
		return false, err
	}
	return reg.MatchString(path), nil
}

func compileCached(pattern string) (*regexp.Regexp, error) {
	relock.Lock()
	defer relock.Unlock()

	regex := regexen[pattern]
	if regex == nil {
		var err error
		regex, err = regexp.Compile("^" + pattern + "$")
		if err != nil {
			return nil, err
		}
		regexen[pattern] = regex
	}
	return regex, nil
}
