package quebec

import (
	"strings"

	"github.com/antchfx/htmlquery"
	"github.com/whitewater-guide/gorge/core"
)

func (s *scriptQuebec) getCodes() ([]string, error) {
	resp, err := core.Client.Get(s.codesURL, &core.RequestOptions{SkipCookies: true})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	doc, err := htmlquery.Parse(resp.Body)
	if err != nil {
		return nil, err
	}
	var result []string
	for _, opt := range htmlquery.Find(doc, "//select[@id='lstStation']/option[@value]") {
		val := htmlquery.SelectAttr(opt, "value")
		result = append(result, strings.TrimSpace(val))
	}
	return result, nil
}
