package uaa

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
)

type UserGuid string

func getUserGuid(i interface{}) (UserGuid, error) {
	s, ok := i.(string)
	if ok != true {
		return "", fmt.Errorf("%v is not a valid user guid", i)
	}

	return UserGuid(s), nil
}

type Client interface {
	LoggedIn() bool
	getAccessToken() (AccessToken, error)
	newHTTPRequest(method, uriStr string, body io.Reader) (*http.Request, error)
	GetServerInfo() (ServerInfo, error)
	ListOauthClients() (OauthClients, error)
	ListIdentityZones() ([]IdentityZone, error)
	ListUsers() (Users, error)
	CreateUser(user *User) (UserGuid, error)
}

type uaaClient struct {
	authenticated bool
	connInfo      *ConnectionInfo
	accessToken   *AccessToken
}

type ConnectionInfo struct {
	ServerURL    string
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

func (c *uaaClient) execute(req *http.Request) (*http.Response, error) {
	if c.accessToken != nil {
		req.Header.Set("Authorization", "Bearer "+c.accessToken.Token)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Failed to submit request: %v", err)
	}

	return resp, err
}

func (c *uaaClient) executeAndUnmarshall(req *http.Request, target interface{}) error {
	resp, err := c.execute(req)
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

	req, err := c.newHTTPRequest("POST", "/oauth/token?grant_type=client_credentials", nil)
	if err != nil {
		return at, err
	}

	req.Header.Add("Accept", "application/json")
	req.SetBasicAuth(c.connInfo.ClientID, c.connInfo.ClientSecret)

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

func (c *uaaClient) CreateUser(user *User) (UserGuid, error) {
	data, err := json.Marshal(user)
	if err != nil {
		return "", err
	}

	requestBody := bytes.NewBuffer(data)

	req, err := c.newHTTPRequest("POST", "/Users", requestBody)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	response, err := c.execute(req)
	if err != nil {
		return "", err
	}

	defer response.Body.Close()
	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("Failed to read response body: %v", err)
	}

	var createUserResponse map[string]interface{}
	err = json.Unmarshal(responseBody, &createUserResponse)
	if err != nil {
		return "", fmt.Errorf("Failed to process POST:/v2/users response: %v", err)
	}

	//fmt.Printf("raw responseBody: %s\ncreateUserResponse: %v\n", string(responseBody), createUserResponse)

	if response.StatusCode == 409 {
		return getUserGuid(createUserResponse["user_id"])
	}

	return getUserGuid(createUserResponse["id"])
}
