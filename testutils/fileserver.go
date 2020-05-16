package testutils

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"text/template"
)

func SetupFileServer(paths map[string]string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// select correct template
		fileTemp := ""
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
		err := t.Execute(&buf, r.URL.Query())
		if err != nil {
			panic("failed to execute template '" + fileTemp + "'")
		}
		filename := buf.String()
		if !strings.HasPrefix(filename, "/") {
			filename = "/" + filename
		}
		filename = "./test_data" + filename

		_, err = os.Stat(filename)
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
