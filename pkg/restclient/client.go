/**
 * Copyright 2019 IBM Corp.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package restclient

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// TODO: implement a real restclient

const (
	GET    = "GET"
	POST   = "POST"
	DELETE = "DELETE"
)

type RestClient interface {
	Get(url string, headers map[string]string, into ...interface{}) Response
	Post(url string, headers map[string]string, body map[string]string, into ...interface{}) Response
	Delete(url string, headers map[string]string) Response
	NewRequest(method, url string, headers, body map[string]string, into ...interface{}) Response
}

type Response struct {
	StatusCode int
	Error      error
	RawContent []byte
}

func (r *Response) IsSuccessful() bool {
	return r.StatusCode >= http.StatusOK && r.StatusCode < http.StatusMultipleChoices
}

func (r *Response) IsUnauthorized() bool {
	return r.StatusCode == http.StatusUnauthorized
}

func (r *Response) IsForbidden() bool {
	return r.StatusCode == http.StatusForbidden
}

type client struct {
	httpClient *http.Client
}

func NewClient() RestClient {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClient := &http.Client{Transport: tr}
	c := &client{httpClient: httpClient}
	return c
}

func (c *client) Get(url string, headers map[string]string, into ...interface{}) Response {
	result := Response{}

	req, err := http.NewRequest(GET, url, nil)
	if err != nil {
		result.Error = err
		return result
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		result.Error = err
		return result
	}

	defer resp.Body.Close()

	c.setContent(resp, &result, into...)

	return result
}

func (c *client) Post(url string, headers map[string]string, body map[string]string, into ...interface{}) Response {
	result := Response{}

	payload, err := json.Marshal(body)
	if err != nil {
		result.Error = err
		return result
	}

	req, err := http.NewRequest(POST, url, bytes.NewBuffer(payload))
	if err != nil {
		result.Error = err
		return result
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		result.Error = err
		return result
	}

	defer resp.Body.Close()

	c.setContent(resp, &result, into...)

	return result
}

func (c *client) Delete(url string, headers map[string]string) Response {
	result := Response{}

	req, err := http.NewRequest(DELETE, url, nil)
	if err != nil {
		result.Error = err
		return result
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		result.Error = err
		return result
	}

	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode

	return result
}

func (c *client) NewRequest(method, url string, headers, body map[string]string, into ...interface{}) Response {
	switch method {
	case GET:
		return c.Get(url, headers, into...)
	case POST:
		return c.Post(url, headers, body, into...)
	case DELETE:
		return c.Delete(url, headers)
	default:
		return Response{}
	}
}

func (c *client) setContent(r *http.Response, result *Response, into ...interface{}) {
	result.StatusCode = r.StatusCode

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		result.Error = err
		return
	} else {
		result.RawContent = body
	}

	if result.IsSuccessful() && len(into) == 1 {
		result.Error = json.Unmarshal(body, into[0])
	}
}
