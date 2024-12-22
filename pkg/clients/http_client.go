package clients

import (
	"errors"
	"io"
	"net/http"
	"time"
)

const timeout = time.Second * 15

var ErrFailedCloseResponseBody = errors.New("failed close response body")

type HTTPClientI interface {
	Do(req *http.Request) (*http.Response, error)
	Get(url string, headers http.Header) (statusCode int, respBody []byte, respHeaders http.Header, err error)
}

type HTTPClientAdapter struct {
	client *http.Client
}

func (h *HTTPClientAdapter) Do(req *http.Request) (*http.Response, error) {
	return h.client.Do(req)
}

func (h *HTTPClientAdapter) Get(url string, headers http.Header) (statusCode int, respBody []byte, respHeaders http.Header, err error) {
	req, err := http.NewRequest(http.MethodGet, url, http.NoBody)
	if err != nil {
		return
	}

	req.Header = headers
	resp, err := h.client.Do(req)
	if err != nil {
		return
	}

	defer func() {
		if e := resp.Body.Close(); e != nil {
			err = errors.Join(err, ErrFailedCloseResponseBody)
		}
	}()

	respBody, err = io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	statusCode = resp.StatusCode
	respHeaders = resp.Header

	return
}

type HTTPClient struct {
	client HTTPClientI
}

func NewHTTPClient() *HTTPClient {
	return &HTTPClient{
		client: &HTTPClientAdapter{
			client: &http.Client{Timeout: timeout},
		},
	}
}

func (h *HTTPClient) Get(url string, headers http.Header) (statusCode int, respBody []byte, respHeaders http.Header, err error) {
	return h.client.Get(url, headers)
}

func (h *HTTPClient) Do(req *http.Request) (*http.Response, error) {
	return h.client.Do(req)
}

func (h *HTTPClient) SetClient(mock HTTPClientI) {
	h.client = mock
}
