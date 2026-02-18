package rmq

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	rabbithole "github.com/michaelklishin/rabbit-hole/v3"
)

type Client struct {
	*rabbithole.Client
	transport http.RoundTripper
	timeout   time.Duration
}

func NewClient(url, username, password string) (*Client, error) {
	client, err := rabbithole.NewClient(url, username, password)
	if err != nil {
		return nil, err
	}

	c := Client{
		Client: client,
	}

	return &c, nil
}

func NewTLSClient(uri, username, password string, transport http.RoundTripper) (*Client, error) {
	client, err := rabbithole.NewTLSClient(uri, username, password, transport)
	if err != nil {
		return nil, err
	}

	c := Client{
		Client:    client,
		transport: transport,
	}

	return &c, nil
}

// SetTimeout changes the HTTP timeout that the Client will use.
// By default, there is no timeout.
func (c *Client) SetTimeout(timeout time.Duration) {
	c.Client.SetTimeout(timeout)
	c.timeout = timeout
}

func newRequestWithBody(client *Client, method string, path string, body []byte) (*http.Request, error) {
	s := client.Endpoint + "/api/" + path

	req, err := http.NewRequest(method, s, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Close = true
	req.SetBasicAuth(client.Username, client.Password)

	req.Header.Add("Content-Type", "application/json")

	return req, err
}

func executeRequest(client *Client, req *http.Request) (resp *http.Response, err error) {
	httpClient := &http.Client{
		Timeout: client.timeout,
	}

	if client.transport != nil {
		httpClient.Transport = client.transport
	}

	resp, err = httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if err = parseResponseErrors(resp); err != nil {
		if resp.Body != nil {
			_ = resp.Body.Close()
		}

		return nil, err
	}

	return resp, err
}

func executeAndParseRequest(client *Client, req *http.Request, rec interface{}) (err error) {
	res, err := executeRequest(client, req)

	if err != nil {
		return err
	}
	defer res.Body.Close()

	if err = json.NewDecoder(res.Body).Decode(&rec); err != nil {
		return err
	}

	return nil
}

func parseResponseErrors(res *http.Response) (err error) {
	if res.StatusCode == http.StatusUnauthorized {
		return errors.New("error: API responded with a 401 Unauthorized")
	}

	// handle a "404 Not Found" response for a DELETE request as a success.
	if res.Request.Method == http.MethodDelete && res.StatusCode == http.StatusNotFound {
		return nil
	}

	if res.StatusCode >= http.StatusBadRequest {
		rme := rabbithole.ErrorResponse{}
		if err = json.NewDecoder(res.Body).Decode(&rme); err != nil {
			rme.Message = fmt.Sprintf("Error %d from RabbitMQ: %s", res.StatusCode, err)
		}
		rme.StatusCode = res.StatusCode
		return rme
	}

	return nil
}
