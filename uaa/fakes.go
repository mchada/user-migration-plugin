package uaa

import (
	"io"
	"net/http"
)

type FakeUaac struct {
	si           ServerInfo
	oauthClients OauthClients
	zones        []IdentityZone
	users        Users
	err          error
}

func (c *FakeUaac) LoggedIn() bool {
	return true
}

func (c *FakeUaac) GetServerInfoReturns(si ServerInfo, err error) {
	c.si = si
	c.err = err
}

func (c *FakeUaac) ListOauthClientsReturns(o OauthClients, err error) {
	c.oauthClients = o
	c.err = err
}

func (c *FakeUaac) ListIdentityZonesReturns(z []IdentityZone, err error) {
	c.zones = z
	c.err = err
}

func (c *FakeUaac) getAccessToken() (AccessToken, error) {
	panic("fakeUaac.getAccessToken should not be called!!")
}
func (c *FakeUaac) newHTTPRequest(method, uriStr string, body io.Reader) (*http.Request, error) {
	panic("fakeUaac.newHTTPRequest should not be called!!")
}

func (c *FakeUaac) GetServerInfo() (ServerInfo, error) {
	return c.si, c.err
}
func (c *FakeUaac) ListOauthClients() (OauthClients, error) {
	return c.oauthClients, c.err
}
func (c *FakeUaac) ListIdentityZones() ([]IdentityZone, error) {
	return c.zones, c.err
}
func (c *FakeUaac) ListUsers() (Users, error) {
	return c.users, c.err
}
