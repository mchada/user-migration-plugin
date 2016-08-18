package uaa

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strings"
)

type Client interface {
	LoggedIn() bool
	getAccessToken() (AccessToken, error)
	newHTTPRequest(method, uriStr string, body io.Reader) (*http.Request, error)
	GetServerInfo() (ServerInfo, error)
	ListOauthClients() (OauthClients, error)
	ListIdentityZones() ([]IdentityZone, error)
	ListUsers() (Users, error)
}

type uaaClient struct {
	authenticated bool
	connInfo      *ConnectionInfo
	accessToken   *AccessToken
}

type ConnectionInfo struct {
	ServerURL    string `required:"true"`
	ClientID     string `required:"true"`
	ClientSecret string `required:"true"`
}

func (connInfo *ConnectionInfo) Connect() (Client, error) {
	c := &uaaClient{
		connInfo: connInfo,
	}

	at, err := c.getAccessToken()
	if err != nil {
		return nil, fmt.Errorf("Failed to get access token: %s", err.Error())
	}

	c.accessToken = &at
	c.authenticated = true

	return c, nil
}

func (c *uaaClient) LoggedIn() bool {
	return c.authenticated
}

func (c *uaaClient) newHTTPRequest(method, uriStr string, body io.Reader) (*http.Request, error) {
	return http.NewRequest(method, c.connInfo.ServerURL+uriStr, body)
}

func (c *uaaClient) executeAndUnmarshall(req *http.Request, target interface{}) error {
	if c.accessToken != nil {
		req.Header.Set("Authorization", "Bearer "+c.accessToken.Token)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Failed to submit request: %v", err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Failed to read response body: %v", err)
	}

	err = json.Unmarshal(body, &target)
	if err != nil {
		return fmt.Errorf("Unable to unmarshall response body to type %s; error: %v; response body: %s", reflect.TypeOf(target), err.Error(), string(body))
	}

	return nil
}

func (c *uaaClient) getAccessToken() (AccessToken, error) {
	var at AccessToken

	params := url.Values{}
	params.Set("client_id", c.connInfo.ClientID)
	params.Set("client_secret", c.connInfo.ClientSecret)
	params.Set("grant_type", "client_credentials")
	params.Set("response_type", "token")

	req, err := c.newHTTPRequest("POST", "/oauth/token", strings.NewReader(params.Encode()))
	if err != nil {
		return at, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	err = c.executeAndUnmarshall(req, &at)
	if err != nil {
		return at, err
	}

	return at, nil
}

func (c *uaaClient) GetServerInfo() (ServerInfo, error) {
	var info ServerInfo

	req, err := c.newHTTPRequest("GET", "/info", nil)
	if err != nil {
		return info, err
	}

	req.Header.Set("Accept", "application/json")
	err = c.executeAndUnmarshall(req, &info)
	if err != nil {
		return info, err
	}

	return info, nil
}

func (c *uaaClient) ListOauthClients() (OauthClients, error) {
	var clients OauthClients

	req, err := c.newHTTPRequest("GET", "/oauth/clients", nil)
	if err != nil {
		return clients, err
	}

	err = c.executeAndUnmarshall(req, &clients)
	if err != nil {
		return clients, err
	}

	return clients, nil
}

func (c *uaaClient) ListIdentityZones() ([]IdentityZone, error) {
	var zones []IdentityZone

	req, err := c.newHTTPRequest("GET", "/identity-zones", nil)
	if err != nil {
		return zones, err
	}

	err = c.executeAndUnmarshall(req, &zones)
	if err != nil {
		return zones, err
	}

	return zones, nil
}

func (c *uaaClient) ListUsers() (Users, error) {
	var users Users

	req, err := c.newHTTPRequest("GET", "/Users", nil)
	if err != nil {
		return users, err
	}

	req.Header.Set("Accept", "application/json")

	err = c.executeAndUnmarshall(req, &users)
	if err != nil {
		return users, err
	}

	return users, nil
}
