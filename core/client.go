package core

import (
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	browser "github.com/EDDYCJY/fake-useragent"
	jar "github.com/juju/persistent-cookiejar"
	log "github.com/sirupsen/logrus"
	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"
)

// HTTPClient is like default client, but with some conveniency methods for common scenarios
type HTTPClient struct {
	*http.Client
	PersistentJar *jar.Jar
	UserAgent     string
}

// ClientOptions are HTTPClient that can be passed as args at startup
type ClientOptions struct {
	UserAgent string `desc:"User agent for requests sent from scripts. Leave empty to use fake browser agent"`
	Timeout   int64  `desc:"Request timeout in seconds"`
}

// RequestOptions are additional per-request options
type RequestOptions struct {
	// When set to true, requests will be sent with random user-agent
	FakeAgent bool
}

// Client is default client for scripts
var Client = NewClient(ClientOptions{
	UserAgent: "whitewater.guide robot",
	Timeout:   60,
})

// NewClient constructs new HTTPClient with options
func NewClient(opts ClientOptions) *HTTPClient {
	jarOpts := jar.Options{
		Filename: "/tmp/cookies/gorge.cookies",
	}
	persJar, err := jar.New(&jarOpts)
	if err != nil {
		log.Fatalf("Failed to initialize cookie jar: %w", err)
		return nil
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DisableKeepAlives = true
	client := &HTTPClient{
		Client:        &http.Client{Jar: persJar, Transport: transport},
		PersistentJar: persJar,
	}

	client.Timeout = time.Duration(opts.Timeout) * time.Second
	client.UserAgent = opts.UserAgent

	return client
}

// EnsureCookie makes sure that cookies from given URL are present and will be sent with further requests
// Some scripts will not return correct data unless cookies are present
func (client *HTTPClient) EnsureCookie(fromURL string, force bool) error {
	cURL, err := url.Parse(fromURL)
	if err != nil {
		return WrapErr(err, "failed to parse cookie URL").With("url", fromURL)
	}
	cookies := client.PersistentJar.Cookies(cURL)
	if force || len(cookies) == 0 {
		resp, err := client.Get(fromURL, nil)
		if err != nil {
			return WrapErr(err, "failed to fetch cookie URL").With("url", fromURL)
		}
		resp.Body.Close()
	}
	return nil
}

// SaveCookies dumps cookies to disk, so in case of service restart they are not lost
func (client *HTTPClient) SaveCookies() {
	client.PersistentJar.Save() //nolint:errcheck
}

// Get is same as http.Client.Get, but sets extra headers and is cached in development environment
func (client *HTTPClient) Get(url string, opts *RequestOptions) (resp *http.Response, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	resp, err = client.Do(req, opts)
	return
}

// Do is same as http.Client.Get, but sets extra headers
func (client *HTTPClient) Do(req *http.Request, opts *RequestOptions) (*http.Response, error) {
	ua := client.UserAgent
	if opts != nil && opts.FakeAgent {
		ua = browser.MacOSX()
	}
	req.Header.Set("User-Agent", ua)
	req.Header.Set("Cache-Control", "no-cache")

	return client.Client.Do(req)
}

// GetAsString is shortcut for http.Client.Get to get response as string
func (client *HTTPClient) GetAsString(url string, opts *RequestOptions) (string, error) {
	resp, err := client.Get(url, opts)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// GetAsJSON is shortcut for http.Client.Get to get response as JSON
func (client *HTTPClient) GetAsJSON(url string, dest interface{}, opts *RequestOptions) error {
	resp, err := client.Get(url, opts)

	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(dest)
}

// GetAsXML is shortcut for http.Client.Get to get response as XML
func (client *HTTPClient) GetAsXML(url string, dest interface{}, opts *RequestOptions) error {
	resp, err := client.Get(url, opts)

	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return xml.NewDecoder(resp.Body).Decode(dest)
}

// PostForm is like http.Client.PostForm but wit extra options
func (client *HTTPClient) PostForm(url string, data url.Values, opts *RequestOptions) (resp *http.Response, req *http.Request, err error) {
	req, err = http.NewRequest("POST", url, strings.NewReader(data.Encode()))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err = client.Do(req, opts)
	return
}

// PostFormAsString shortcut for http.Client.PostForm to get response as string
func (client *HTTPClient) PostFormAsString(url string, data url.Values, opts *RequestOptions) (result string, req *http.Request, err error) {
	resp, req, err := client.PostForm(url, data, opts)
	if err != nil {
		return "", req, err
	}
	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", req, err
	}
	return string(bytes), req, nil
}

// CSVStreamOptions contains commons options for streaming data from CSV files
type CSVStreamOptions struct {
	// CSV separator symbol
	Comma rune
	// Decoder, defaults to UTF-8
	Decoder *encoding.Decoder
	// Number of rows at the beginning of file that do not contain data
	HeaderHeight int
	// Number of colums. If a row contains different number of columns, the stream will stop with error
	// For header rows this is ignored
	NumColumns int
	// Extra HTTPClient options
	*RequestOptions
}

// StreamCSV reads CSV file from given URL and streams it by calling handler for each row
func (client *HTTPClient) StreamCSV(url string, handler func(row []string) error, opts CSVStreamOptions) error {
	resp, err := client.Get(url, opts.RequestOptions)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	defer io.Copy(ioutil.Discard, resp.Body) //nolint:errcheck
	var reader io.Reader = resp.Body
	if opts.Decoder != nil {
		reader = transform.NewReader(resp.Body, opts.Decoder)
	}
	csvReader := csv.NewReader(reader)
	csvReader.ReuseRecord = true
	csvReader.FieldsPerRecord = opts.NumColumns
	if opts.Comma != 0 {
		csvReader.Comma = opts.Comma
	}
	skippedHeader := opts.HeaderHeight
	var row []string
	for {
		row, err = csvReader.Read()
		if err == io.EOF {
			break
		} else if e, ok := err.(*csv.ParseError); ok && e.Err == csv.ErrFieldCount {
			skippedHeader--
			continue
		} else if err != nil {
			return WrapErr(err, "csv stream error")
		}
		if skippedHeader > 0 {
			skippedHeader--
			continue
		}
		if opts.NumColumns != 0 && len(row) != opts.NumColumns {
			return NewErr(fmt.Errorf("unexpected csv row with %d columns insteas of %d", len(row), opts.NumColumns)).With("row", row)
		}
		if err = handler(row); err != nil {
			return err
		}
	}
	return nil
}
