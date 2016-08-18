package uaa

import (
	"encoding/json"
	"io/ioutil"
	"testing"
)

func TestJSONUnmarshallAccessToken(t *testing.T) {
	responseBody := []byte(`{
      "access_token":"dummy-access-token",
      "token_type":"dummy-token-type",
      "expires_in":3600,
      "scope":"dummy-scope",
      "jti":"dummy-jti"
    }`)

	var token AccessToken

	err := json.Unmarshal(responseBody, &token)
	if err != nil {
		t.Errorf("Failed to unmarshall json to AccessToken: %v", err)
	}

	if len(token.Token) == 0 {
		t.Errorf("AccessToken.Token field empty after unmarshalling from json")
	}

	if len(token.Type) == 0 {
		t.Errorf("AccessToken.Type field empty after unmarshalling from json")
	}

	if token.ExpiresIn == 0 {
		t.Errorf("AccessToken.ExpiresIn field empty after unmarshalling from json")
	}

	if len(token.Scope) == 0 {
		t.Errorf("AccessToken.Scope field empty after unmarshalling from json")
	}

	if len(token.JTI) == 0 {
		t.Errorf("AccessToken.JTI field empty after unmarshalling from json")
	}
}

func TestJSONUnmarshallOauthClient(t *testing.T) {
	responseBody := []byte(`{
      "client_id" : "foo",
      "name" : "Foo Client Name",
      "scope" : ["uaa.none"],
      "resource_ids" : ["none"],
      "authorities" : ["cloud_controller.read","cloud_controller.write","scim.read"],
      "authorized_grant_types" : ["client_credentials"],
      "lastModified" : 1426260091149
    }`)

	var client OauthClient

	err := json.Unmarshal(responseBody, &client)
	if err != nil {
		t.Errorf("Failed to unmarshall json to OauthClient: %v", err)
	}

	if len(client.ID) == 0 {
		t.Errorf("client.ID field empty after unmarshalling from json")
	}
	if len(client.Name) == 0 {
		t.Errorf("client.Name field empty after unmarshalling from json")
	}

	if len(client.Scope) == 0 {
		t.Errorf("client.Scope field empty after unmarshalling from json")
	}
	if len(client.ResourceIDs) == 0 {
		t.Errorf("client.ResourceIDs field empty after unmarshalling from json")
	}
	if len(client.Authorities) == 0 {
		t.Errorf("client.Authorities field empty after unmarshalling from json")
	}
	if len(client.AuthorizedGrantTypes) == 0 {
		t.Errorf("client.AuthorizedGrantTypes field empty after unmarshalling from json")
	}
	if client.LastModified == 0 {
		t.Errorf("client.LastModified field empty after unmarshalling from json")
	}
}

func TestJSONUnmarshallOauthClients(t *testing.T) {
	responseBody := []byte(`{
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

	var clients OauthClients

	err := json.Unmarshal(responseBody, &clients)
	if err != nil {
		t.Errorf("Failed to unmarshall json to OauthClients: %v", err)
	}
}

func TestJSONUnmarshallServerInfo(t *testing.T) {
	responseBody := []byte(`{
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

	var info ServerInfo

	err := json.Unmarshal(responseBody, &info)
	if err != nil {
		t.Errorf("Failed to unmarshall json to ServerInfo: %v", err)
	}

	if len(info.Version()) == 0 {
		t.Error("Failed to unmarshall app.version to ServerInfo.Version")
	}
}

func TestJSONUnmarshallIdentityZone(t *testing.T) {
	responseBody := []byte(`{
        "id": "dummy-id",
        "subdomain": "dummy-subdomain",
        "name": "dummy-name",
        "version": 1,
        "description": "Dummy Description",
        "created": 946710000000,
        "last_modified": 946710000000
    }`)

	var zone IdentityZone

	err := json.Unmarshal(responseBody, &zone)
	if err != nil {
		t.Errorf("Failed to unmarshall json to Zone: %v", err)
	}

	if len(zone.ID) == 0 {
		t.Errorf("Zone.ID field empty after unmarshalling from json")
	}
	if len(zone.Subdomain) == 0 {
		t.Errorf("Zone.Subdomain field empty after unmarshalling from json")
	}
	if len(zone.Name) == 0 {
		t.Errorf("Zone.Name field empty after unmarshalling from json")
	}
	if len(zone.Description) == 0 {
		t.Errorf("Zone.Description field empty after unmarshalling from json")
	}
	if zone.Version == 0 {
		t.Errorf("Zone.Version field empty after unmarshalling from json")
	}
	if zone.Created == 0 {
		t.Errorf("Zone.Created field empty after unmarshalling from json")
	}
	if zone.LastModified == 0 {
		t.Errorf("Zone.LastModified field empty after unmarshalling from json")
	}
}

func TestJSONUnmarshallUsers(t *testing.T) {
	responseBody, err := ioutil.ReadFile("../testdata/users.json")

	if err != nil {
		panic("Failed to read ../testdata/users.json: " + err.Error())
	}

	var users Users

	err = json.Unmarshal(responseBody, &users)
	if err != nil {
		t.Errorf("Failed to unmarshall json to Users: %v", err)
	}

	if len(users.Users) == 0 {
		t.Error("Failed to unmarshall resources json field onto Users.Users")
	}

	user := users.Users[0]

	if len(user.GUID) == 0 {
		t.Error("Failed to unmarshall resources[0].id field onto Users.Users[0].GUID")
	}
	if len(user.ExternalID) == 0 {
		t.Error("Failed to unmarshall resources[0].externalID field onto Users.Users[0].ExternalID")
	}
	if len(user.Username) == 0 {
		t.Error("Failed to unmarshall resources[0].username field onto Users.Users[0].Username")
	}
}

func readTestdata(filename string) ([]byte, error) {
	return ioutil.ReadFile(filename)
}
