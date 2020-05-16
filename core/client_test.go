package core

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/text/encoding/charmap"
)

type testJSON struct {
	Foo string `json:"foo"`
	Baz int64  `json:"baz"`
}

func TestHttpClient_Get(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if r.Header.Get("Cache-Control") != "no-cache" {
			t.Errorf("Expected cache-control: no-cache")
		}
		if r.Header.Get("User-Agent") != "whitewater.guide robot" {
			t.Errorf("Expected User-Agent: whitewater.guide robot")
		}
	}))
	defer ts.Close()
	_, _ = Client.Get(ts.URL, nil)
}

func TestHttpClient_GetAsJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"foo": "bar", "baz": 42}`)
	}))
	defer ts.Close()
	actual := &testJSON{}
	expected := &testJSON{Foo: "bar", Baz: 42}
	err := Client.GetAsJSON(ts.URL, actual, nil)
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}

func TestHttpClient_GetAsDoc(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `<foo><bar>baz</bar></foo>`)
	}))
	defer ts.Close()
	doc, err := Client.GetAsDoc(ts.URL, nil)
	if assert.NoError(t, err) {
		assert.Equal(t, "baz", doc.Find("bar").Text())
		doc.Close()
		_, err = doc.resp.Body.Read(nil)
		assert.True(t, err != nil && err.Error() == "http: read on closed response body")
	}
}

func TestHttpClient_GetFakeAgent(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if r.Header.Get("Cache-Control") != "no-cache" {
			t.Errorf("Expected cache-control: no-cache")
		}
		ua := r.Header.Get("User-Agent")
		if ua == "whitewater.guide robot" {
			t.Errorf("Expected fake user-agent, got '%s'", ua)
		}
	}))
	defer ts.Close()
	_, _ = Client.Get(ts.URL, &RequestOptions{FakeAgent: true})
}

func TestHttpClient_StreamCSV(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Add("Content-Type", "text/csw")
		fmt.Fprint(w, "h1,h2,h3\nv1,v2,ú\nx1,x2,x3")
	}))
	defer ts.Close()
	// error in handler
	tests := []struct {
		name       string
		opts       CSVStreamOptions
		handlerErr error
		calls      [][]string
		err        bool
	}{
		{
			name:  "default",
			opts:  CSVStreamOptions{HeaderHeight: 1},
			calls: [][]string{{"v1", "v2", "ú"}, {"x1", "x2", "x3"}},
		},
		{
			name:  "default with explicit options",
			opts:  CSVStreamOptions{HeaderHeight: 1, NumColumns: 3, Comma: ','},
			calls: [][]string{{"v1", "v2", "ú"}, {"x1", "x2", "x3"}},
		},
		{
			name:  "no header",
			opts:  CSVStreamOptions{HeaderHeight: 0},
			calls: [][]string{{"h1", "h2", "h3"}, {"v1", "v2", "ú"}, {"x1", "x2", "x3"}},
		},
		{
			name:  "long header",
			opts:  CSVStreamOptions{HeaderHeight: 2},
			calls: [][]string{{"x1", "x2", "x3"}},
		},
		{
			name:  "windows-1251 encoding",
			opts:  CSVStreamOptions{HeaderHeight: 1, Decoder: charmap.Windows1251.NewDecoder()},
			calls: [][]string{{"v1", "v2", "Гє"}, {"x1", "x2", "x3"}},
		},
		{
			name:       "error in handler",
			handlerErr: errors.New("boom"),
			opts:       CSVStreamOptions{HeaderHeight: 1},
			err:        true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var actual [][]string
			sink := func(row []string) error {
				if tt.handlerErr != nil {
					return tt.handlerErr
				}
				ar := append(row[:0:0], row...)
				actual = append(actual, ar)
				return nil
			}
			err := Client.StreamCSV(ts.URL, sink, tt.opts)
			if tt.err {
				assert.Error(t, err)
			} else if assert.NoError(t, err) {
				assert.Equal(t, tt.calls, actual)
			}
		})
	}
}
