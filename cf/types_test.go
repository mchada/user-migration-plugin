package cf

import (
	"encoding/json"
	"io/ioutil"
	"testing"
)

func TestJSONUnmarshallUserWith243Model(t *testing.T) {
	responseBody := []byte(`{
     "admin": false,
     "active": true,
     "default_space_guid": null,
     "username": "push_apps_manager",
     "spaces_url": "/v2/users/96c1d7f1-0f3e-4946-a041-003cea7650eb/spaces",
     "organizations_url": "/v2/users/96c1d7f1-0f3e-4946-a041-003cea7650eb/organizations",
     "managed_organizations_url": "/v2/users/96c1d7f1-0f3e-4946-a041-003cea7650eb/managed_organizations",
     "billing_managed_organizations_url": "/v2/users/96c1d7f1-0f3e-4946-a041-003cea7650eb/billing_managed_organizations",
     "audited_organizations_url": "/v2/users/96c1d7f1-0f3e-4946-a041-003cea7650eb/audited_organizations",
     "managed_spaces_url": "/v2/users/96c1d7f1-0f3e-4946-a041-003cea7650eb/managed_spaces",
     "audited_spaces_url": "/v2/users/96c1d7f1-0f3e-4946-a041-003cea7650eb/audited_spaces"
  }`)

	var user User

	err := json.Unmarshal(responseBody, &user)
	if err != nil {
		t.Errorf("Failed to unmarshall json to User: %v", err)
	}

	if len(user.Username) == 0 {
		t.Error("Failed to unmarshall Username field")
	}
}

func TestJSONUnmarshallUserResourceWith243Model(t *testing.T) {
	responseBody := []byte(`{
     "metadata": {
        "guid": "96c1d7f1-0f3e-4946-a041-003cea7650eb",
        "url": "/v2/users/96c1d7f1-0f3e-4946-a041-003cea7650eb",
        "created_at": "2015-10-13T03:31:34Z",
        "updated_at": null
     },
     "entity": {
        "admin": false,
        "active": true,
        "default_space_guid": null,
        "username": "push_apps_manager",
        "spaces_url": "/v2/users/96c1d7f1-0f3e-4946-a041-003cea7650eb/spaces",
        "organizations_url": "/v2/users/96c1d7f1-0f3e-4946-a041-003cea7650eb/organizations",
        "managed_organizations_url": "/v2/users/96c1d7f1-0f3e-4946-a041-003cea7650eb/managed_organizations",
        "billing_managed_organizations_url": "/v2/users/96c1d7f1-0f3e-4946-a041-003cea7650eb/billing_managed_organizations",
        "audited_organizations_url": "/v2/users/96c1d7f1-0f3e-4946-a041-003cea7650eb/audited_organizations",
        "managed_spaces_url": "/v2/users/96c1d7f1-0f3e-4946-a041-003cea7650eb/managed_spaces",
        "audited_spaces_url": "/v2/users/96c1d7f1-0f3e-4946-a041-003cea7650eb/audited_spaces"
     }
  }`)

	var userResource UserResource

	err := json.Unmarshal(responseBody, &userResource)
	if err != nil {
		t.Errorf("Failed to unmarshall json to userResource: %v", err)
	}

	if len(userResource.Entity.Username) == 0 {
		t.Error("Failed to unmarshall userResource.Entity.Username field")
	}

	if len(userResource.Metadata.GUID) == 0 {
		t.Error("Failed to unmarshall userResource.Metadata.GUID field")
	}
}

func TestJSONUnmarshallUsersResponseWith243Model(t *testing.T) {
	responseBody, err := ioutil.ReadFile("./testdata/cc-list-users-2.43.0.json")

	if err != nil {
		panic("Failed to read ../testdata/cc-list-users-2.43.0.json: " + err.Error())
	}

	var usersResponse UsersResponse

	err = json.Unmarshal(responseBody, &usersResponse)
	if err != nil {
		t.Errorf("Failed to unmarshall json to usersResponse: %v", err)
	}

	if len(usersResponse.Resources) == 0 {
		t.Error("Failed to unmarshall resources json field onto usersResponse.Resources")
	}

	for _, userResource := range usersResponse.Resources {
		if len(userResource.Metadata.GUID) == 0 {
			t.Error("Failed to unmarshall userResource.Metadata.GUID field")
		}
	}
}
