package core

import (
	"crypto/tls"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"moul.io/http2curl/v2"

	browser "github.com/EDDYCJY/fake-useragent"
	"github.com/PuerkitoBio/goquery"
	"github.com/cenkalti/backoff/v4"
	jar "github.com/juju/persistent-cookiejar"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"
)

// HTTPClient is like default client, but with some conveniency methods for common scenarios
type HTTPClient struct {
	*http.Client
	PersistentJar *jar.Jar
	UserAgent     string
	logger        *logrus.Entry
}

// ClientOptions are HTTPClient that can be passed as args at startup
type ClientOptions struct {
	UserAgent  string `desc:"User agent for requests sent from scripts. Leave empty to use fake browser agent"`
	Timeout    int64  `desc:"Request timeout in seconds"`
	WithoutTLS bool   `desc:"Disable TLS for some gauges"`
	Proxy      string `desc:"HTTP client proxy (for example, you can use mitm for local development)"`
}

// RequestOptions are additional per-request options
type RequestOptions struct {
	// When set to true, requests will be sent with random user-agent
	FakeAgent bool
	// Headers to set on request
	Headers map[string]string
	// Request will not save cookies
	SkipCookies bool
	// Retry request which returned >= 400 status code
	RetryErrors bool
}

// Client is default client for scripts
// It will be reinitialized during server creation
// This default value will be used in tests
var Client = NewClient(ClientOptions{
	UserAgent:  "whitewater.guide robot",
	Timeout:    60,
	WithoutTLS: false,
	Proxy:      "",
}, nil)

// NewClient constructs new HTTPClient with options
func NewClient(opts ClientOptions, logger *logrus.Entry) *HTTPClient {
	jarOpts := jar.Options{
		Filename: "/tmp/cookies/gorge.cookies",
	}
	persJar, err := jar.New(&jarOpts)
	if err != nil {
		log.Fatalf("Failed to initialize cookie jar: %v", err)
		return nil
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DisableKeepAlives = true
	if proxy, perr := url.Parse(opts.Proxy); perr == nil && opts.Proxy != "" {
		transport.Proxy = func(req *http.Request) (*url.URL, error) {
			return proxy, nil
		}
	}
	if opts.WithoutTLS {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	client := &HTTPClient{
		Client:        &http.Client{Jar: persJar, Transport: transport},
		PersistentJar: persJar,
		logger:        logger,
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

// Do is same as http.Client.Get, but sets extra headers
func (client *HTTPClient) Do(req *http.Request, opts *RequestOptions) (*http.Response, error) {
	retryErrors := false
	ua := client.UserAgent
	if opts != nil && opts.FakeAgent {
		ua = browser.MacOSX()
	}
	req.Header.Set("User-Agent", ua)
	req.Header.Set("Cache-Control", "no-cache")
	if opts != nil {
		for k, v := range opts.Headers {
			req.Header.Set(k, v)
		}
		retryErrors = opts.RetryErrors
	}

	if client.logger != nil {
		client.logger.Debug(http2curl.GetCurlCommand(req))
	}

	resp, err := client.doRetry(req, retryErrors)

	if opts != nil && resp != nil && opts.SkipCookies {
		cookies := resp.Cookies()
		for _, rc := range cookies {
			rc.MaxAge = -1
			client.Jar.SetCookies(resp.Request.URL, []*http.Cookie{rc})
		}
	}

	return resp, err
}

func (client *HTTPClient) doRetry(req *http.Request, retryErrors bool) (*http.Response, error) {
	b := backoff.WithMaxRetries(backoff.NewExponentialBackOff(backoff.WithInitialInterval(time.Second)), 3)

	if retryErrors {
		return backoff.RetryWithData(func() (*http.Response, error) {
			resp, err := client.Client.Do(req)
			if err == nil && resp.StatusCode >= 400 {
				err = fmt.Errorf("req failed (%d): %s", resp.StatusCode, resp.Status)
			}
			return resp, err
		}, b)

	}
	return client.Client.Do(req)
}

// Get is same as http.Client.Get, but sets extra headers
func (client *HTTPClient) Get(url string, opts *RequestOptions) (resp *http.Response, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	resp, err = client.Do(req, opts)
	return
}

// GetAsString is shortcut for http.Client.Get to get response as string
func (client *HTTPClient) GetAsString(url string, opts *RequestOptions) (string, error) {
	resp, err := client.Get(url, opts)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	bytes, err := io.ReadAll(resp.Body)
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

// Doc extends goquery.Document with a Close() method
type Doc struct {
	*goquery.Document
	resp *http.Response
}

// Close closes underlying resp body
func (doc *Doc) Close() {
	doc.resp.Body.Close()
}

// GetAsDoc is shortcut for http.Client.Get to get HTML docs for goquery.
func (client *HTTPClient) GetAsDoc(url string, opts *RequestOptions) (*Doc, error) {
	resp, err := client.Get(url, opts)
	if err != nil {
		return nil, err
	}

	// Sometimes this return document with empty body
	qdoc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}
	return &Doc{
		Document: qdoc,
		resp:     resp,
	}, nil
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
	bytes, err := io.ReadAll(resp.Body)
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
	defer io.Copy(io.Discard, resp.Body) //nolint:errcheck
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
