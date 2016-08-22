package uaa

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{
      "access_token" : "4a53a3331b2445cfaca43c9af00439e8",
      "token_type" : "bearer",
      "expires_in" : 43199,
      "scope" : "clients.read emails.write scim.userids password.write idps.write notifications.write oauth.login scim.write critical_notifications.write",
      "jti" : "4a53a3331b2445cfaca43c9af00439e8"
    }`)
	}))
	defer ts.Close()

	connInfo := &ConnectionInfo{
		ServerURL:    ts.URL,
		ClientID:     "fake-client",
		ClientSecret: "big-secret",
	}

	if _, err := connInfo.Connect(); err != nil {
		t.Errorf("Failed to initialize client: %s", err.Error())
	}
}

func TestGetAccessToken(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{
      "access_token" : "4a53a3331b2445cfaca43c9af00439e8",
      "token_type" : "bearer",
      "expires_in" : 43199,
      "scope" : "clients.read emails.write scim.userids password.write idps.write notifications.write oauth.login scim.write critical_notifications.write",
      "jti" : "4a53a3331b2445cfaca43c9af00439e8"
    }`)
	}))
	defer ts.Close()

	connInfo := &ConnectionInfo{
		ServerURL:    ts.URL,
		ClientID:     "fake-client",
		ClientSecret: "big-secret",
	}

	uaac := &uaaClient{
		connInfo: connInfo,
	}

	accessToken, err := uaac.getAccessToken()
	if err != nil {
		t.Errorf("Failed to get access token: %s", err.Error())
	}

	if len(accessToken.Token) == 0 {
		t.Error("AccessToken.Token was blank")
	}
}

func TestGetServerInfo(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{
      "app" : {
        "version" : "3.5.0-SNAPSHOT"
      },
      "links" : {
        "uaa" : "http://localhost:8080/uaa",
        "passwd" : "/forgot_password",
        "login" : "http://localhost:8080/uaa",
        "register" : "/create_account"
      },
      "zone_name" : "uaa",
      "entityID" : "cloudfoundry-saml-login",
      "commit_id" : "6681d65",
      "idpDefinitions" : {
        "SAML" : "http://localhost:8080/uaa/saml/discovery?returnIDParam=idp&entityID=cloudfoundry-saml-login&idp=SAML&isPassive=true"
      },
      "prompts" : {
        "username" : [ "text", "Email" ],
        "password" : [ "password", "Password" ]
      },
      "timestamp" : "2016-05-18T18:20:54+0000"
    }`)
	}))
	defer ts.Close()

	connInfo := &ConnectionInfo{
		ServerURL:    ts.URL,
		ClientID:     "fake-client",
		ClientSecret: "big-secret",
	}

	uaac := &uaaClient{
		connInfo: connInfo,
	}

	serverInfo, err := uaac.GetServerInfo()
	if err != nil {
		t.Errorf("Failed to get ServerInfo: %v", err)
	}

	if len(serverInfo.Version()) == 0 {
		t.Error("ServerInfo.Version() was blank")
	}
}

func TestGetListOauthClients(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{
      "resources" : [ {
        "scope" : [ "clients.read", "clients.write" ],
        "client_id" : "vleI0g",
        "resource_ids" : [ "none" ],
        "authorized_grant_types" : [ "client_credentials" ],
        "redirect_uri" : [ "http*://ant.path.wildcard/**/passback/*", "http://test1.com" ],
        "autoapprove" : [ ],
        "authorities" : [ "clients.read", "clients.write" ],
        "lastModified" : 1463595914005
      } ],
      "startIndex" : 1,
      "itemsPerPage" : 1,
      "totalResults" : 1,
      "schemas" : [ "http://cloudfoundry.org/schema/scim/oauth-clients-1.0" ]
    }`)
	}))
	defer ts.Close()

	connInfo := &ConnectionInfo{
		ServerURL:    ts.URL,
		ClientID:     "fake-client",
		ClientSecret: "big-secret",
	}

	uaac := &uaaClient{
		connInfo: connInfo,
	}

	clients, err := uaac.ListOauthClients()
	if err != nil {
		t.Errorf("Failed to get OauthClients: %v", err)
	}

	if len(clients.Clients) == 0 {
		t.Error("[]OauthClient was empty")
	}
}

func TestGetListIdentityZones(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `[{
        "id": "dummy-id",
        "subdomain": "dummy-subdomain",
        "name": "dummy-name",
        "version": 1,
        "description": "Dummy Description",
        "created": 946710000000,
        "last_modified": 946710000000
    }]`)
	}))
	defer ts.Close()

	connInfo := &ConnectionInfo{
		ServerURL:    ts.URL,
		ClientID:     "fake-client",
		ClientSecret: "big-secret",
	}

	uaac := &uaaClient{
		connInfo: connInfo,
	}

	zones, err := uaac.ListIdentityZones()
	if err != nil {
		t.Errorf("Failed to get Zones: %v", err)
	}

	if len(zones) == 0 {
		t.Error("[]Zone was empty")
	}
}

func TestGetListUsersWithUaa20Model(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := ioutil.ReadFile("../test-data/list-users-2.0.json")

		if err != nil {
			panic("Failed to read ../test-data/list-users-2.0.json: " + err.Error())
		}

		w.Write(data)
	}))
	defer ts.Close()

	connInfo := &ConnectionInfo{
		ServerURL:    ts.URL,
		ClientID:     "fake-client",
		ClientSecret: "big-secret",
	}

	uaac := &uaaClient{
		connInfo: connInfo,
	}

	users, err := uaac.ListUsers()
	if err != nil {
		t.Errorf("Failed to get Users: %v", err)
	}

	if len(users.Users) == 0 {
		t.Error("[]Users was empty")
	}
}
