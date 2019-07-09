package arrayactions

import (
	"fmt"
	"strings"

	"github.com/IBM/ibm-block-csi-driver-operator/pkg/restclient"
	"github.com/go-logr/logr"
)

const svcPort = "7443"
const baseUrl = "rest"

type Token struct {
	Token string
}

type svcMediator struct {
	arrayAddress, username, password, token string
	logger                                  logr.Logger
	client                                  restclient.RestClient
}

func NewSvcMediator(arrayAddress, username, password string, logger logr.Logger) ArrayMediator {
	client := restclient.NewClient()

	return &svcMediator{
		arrayAddress: arrayAddress,
		username:     username,
		password:     password,
		logger:       logger,
		client:       client,
	}
}

func (m *svcMediator) CreateHost(name string, iscsiPorts, fcPorts []string) error {
	endpoint := "mkhost"

	m.logger.Info("Creating host", "name", name, "ports", append(iscsiPorts, fcPorts...))

	url := m.buildUrl(endpoint)
	req_body := map[string]string{
		"name": name,
		"type": "generic",
	}
	if len(fcPorts) > 0 {
		req_body["fcwwpn"] = strings.Join(fcPorts, ",")
	} else {
		req_body["iscsiname"] = strings.Join(iscsiPorts, ",")
	}

	err := m.callAndRetryIfUnauthorized(restclient.POST, url, req_body)
	if err != nil {
		m.logger.Error(err, "Failed to create host", "name", name)
		return err
	}
	m.logger.Info("Created host", "name", name)
	return nil
}

func (m *svcMediator) callAndRetryIfUnauthorized(method, url string, body map[string]string, into ...interface{}) error {
	headers := m.getHeaders()
	authorized, err := m.call(method, url, headers, body, into...)
	if !authorized {
		err := m.setToken()
		if err != nil {
			return err
		}

		// get headers again after set token
		headers = m.getHeaders()
		_, err = m.call(method, url, headers, body, into...)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

func (m *svcMediator) call(method, url string, headers, body map[string]string, into ...interface{}) (bool, error) {
	m.logger.Info(
		"Calling request %s %s, header is %v, body is %v",
		"method", method,
		"url", url,
		"headers", headers,
		"body", body)

	var authorized bool = true
	res := m.client.NewRequest(method, url, headers, body, into...)
	if res.Error != nil {
		return authorized, res.Error
	} else {
		if !res.IsSuccessful() {
			err := fmt.Errorf("Request failed, status code is %d, reason is %s", res.StatusCode, res.RawContent)

			// should be Unauthorized, but svc is stupid, it returns Forbidden
			if res.IsForbidden() {
				authorized = false
			}
			return authorized, err
		}
	}
	return authorized, nil
}

func (m *svcMediator) setToken() error {
	endpoint := "auth"
	m.clearToken()

	url := m.buildUrl(endpoint)
	t := Token{}
	headers := m.getHeaders()
	headers["X-Auth-Username"] = m.username
	headers["X-Auth-Password"] = m.password
	res := m.client.Post(url, headers, nil, &t)
	if res.Error != nil {
		return res.Error
	} else if !res.IsSuccessful() {
		err := fmt.Errorf("Failed to get token, status code is %d, reason is %s", res.StatusCode, res.RawContent)
		m.logger.Error(err, "")
		return err
	}
	m.token = t.Token
	return nil
}

func (m *svcMediator) clearToken() {
	m.token = ""
}

func (m *svcMediator) buildUrl(endpoint string) string {
	return fmt.Sprintf("https://%s:%s/%s/%s", m.arrayAddress, svcPort, baseUrl, endpoint)
}

func (m *svcMediator) getHeaders() map[string]string {
	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	headers["Accept"] = "application/json"
	if m.token != "" {
		headers["X-Auth-Token"] = m.token
	}
	return headers
}
