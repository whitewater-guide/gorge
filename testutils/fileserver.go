package testutils

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"text/template"
)

// Authorizer is used in test server to simulate auth process
type Authorizer interface {
	Authorize(req *http.Request) bool
}

// TestAuthKey is used by HeaderAuthorizer
const TestAuthKey = "__the_key__"

// HeaderAuthorizer checks that value of "Key" header equal TestAuthKey
type HeaderAuthorizer struct {
	// Key is the name of authorization header
	Key string
}

// Authorize implements Authorizer interface
func (a *HeaderAuthorizer) Authorize(req *http.Request) bool {
	key := req.Header.Get(a.Key)
	return key == TestAuthKey
}

// SetupFileServer serves files from './test_data' directory in package
// paths map can be used to servre files based on URL path prefixes and query params (for example, check usgs)
// if paths is nil, url path will be used to find files in dir
func SetupFileServer(paths map[string]string, auth Authorizer) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if auth != nil {
			ok := auth.Authorize(r)
			if !ok {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Unauthorized.\n")) //nolint:errcheck
				return
			}
		}
		fmt.Println("filename", r.URL)
		filename := "./test_data" + r.URL.Path
		fileTemp := ""
		if paths != nil {
			// select correct template
			for prefix, pathTemp := range paths {
				if !strings.HasPrefix(prefix, "/") {
					prefix = "/" + prefix
				}
				if prefix == "/" {
					fileTemp = pathTemp
					continue
				}
				if strings.HasPrefix(r.URL.Path, prefix) {
					fileTemp = pathTemp
					break
				}
			}
			if fileTemp == "" {
				panic("template not found for '" + r.URL.String() + "'")
			}

			var buf bytes.Buffer
			t := template.Must(template.New("").Parse(fileTemp))

			// URL query has []string as values, they is joined
			// URL can have $ symbol in query vars, this should be removed
			tempData := map[string]string{}
			for k, vals := range r.URL.Query() {
				key := strings.Replace(k, "$", "", -1)
				tempData[key] = strings.Join(vals, ",")
			}

			err := t.Execute(&buf, tempData)
			if err != nil {
				panic("failed to execute template '" + fileTemp + "'")
			}
			filename = buf.String()
			if !strings.HasPrefix(filename, "/") {
				filename = "/" + filename
			}
			filename = "./test_data" + filename
		}

		_, err := os.Stat(filename)
		if os.IsNotExist(err) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		file, _ := os.Open(filename)
		w.WriteHeader(http.StatusOK)
		_, err = io.Copy(w, file)
		if err != nil {
			panic("failed to send test file '" + filename + "' from template '" + fileTemp + "'")
		}
	}))
}
