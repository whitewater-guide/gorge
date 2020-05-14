package nz_bop

import (
	"regexp"

	"github.com/whitewater-guide/gorge/core"
)

var codesRegExp = regexp.MustCompile(`d\.add\(\d+,558,.*site=(\d+)`)

func (s *scriptBop) parseList() ([]string, error) {
	html, err := core.Client.GetAsString(s.listURL, nil)
	if err != nil {
		return nil, err
	}
	matches := codesRegExp.FindAllStringSubmatch(html, -1)
	codes := make([]string, len(matches))
	for i, m := range matches {
		codes[i] = m[1]
	}
	return codes, nil
}
