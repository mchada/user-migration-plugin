package cf

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/cloudfoundry/cli/plugin"
	"github.com/pivotalservices/user-migration-plugin/uaa"
)

type Client struct {
	orgsByName   map[string]*OrgResource
	spacesByName map[string]*SpaceResource
}

func NewClient() *Client {
	return &Client{
		orgsByName:   make(map[string]*OrgResource, 0),
		spacesByName: make(map[string]*SpaceResource, 0),
	}
}

func (c *Client) GetUsers(cli plugin.CliConnection) (UsersResponse, error) {
	var usersResponse UsersResponse
	data, err := c.cfcurl(cli, "/v2/users?results-per-page=100")
	if nil != err {
		return usersResponse, err
	}

	err = json.Unmarshal(data, &usersResponse)
	if nil != err {
		fmt.Printf("Failed to parse json get users in org json: %v\njson: %s\n", err, string(data))
		return usersResponse, err
	}

	return usersResponse, err
}

func (c *Client) CreateUser(cli plugin.CliConnection, guid uaa.UserGuid) error {
	payload := fmt.Sprintf(`{"guid": "%s"}`, guid)
	_, err := c.cfcurl(cli, "/v2/users", "-X", "POST", "-d", payload)
	if nil != err {
		return err
	}

	return nil
}

func (c *Client) GetUserSummary(cli plugin.CliConnection, userResource *UserResource) (UserSummaryResource, error) {
	var userSummaryResource UserSummaryResource

	data, err := c.cfcurl(cli, fmt.Sprintf("/v2/users/%s/summary", userResource.Metadata.GUID))

	if nil != err {
		fmt.Println("failed to get summary from server: ", err.Error())
		return userSummaryResource, err
	}

	err = json.Unmarshal(data, &userSummaryResource)

	if nil != err {
		fmt.Println("Failed to parse json: ", err.Error())
		return userSummaryResource, err
	}

	return userSummaryResource, err
}

func (c *Client) SetOrgRoles(cli plugin.CliConnection, username string, orgRoles []*OrgRole) error {
	for _, orgRole := range orgRoles {
		orgResource, err := c.GetOrgByName(cli, orgRole.OrgName)
		if err != nil {
			log.Printf("%v", err)
			continue
		}

		orgGuid := orgResource.Metadata.GUID

		if err := c.AssociateUserWithOrg(cli, orgGuid, username); err != nil {
			fmt.Printf("failed to associate user %s with org %s; %v", username, orgRole.OrgName, err)
		}

		if err := c.SetOrgRole(cli, orgGuid, username, orgRole.RoleName); err != nil {
			fmt.Printf("failed to set org role; user %s - org %s - org role: %s; %v", username, orgRole.OrgName, orgRole.RoleName, err)
		}
	}

	return nil
}

func (c *Client) AssociateUserWithOrg(cli plugin.CliConnection, orgGuid string, username string) error {
	uri := fmt.Sprintf("/v2/organizations/%s/users", orgGuid)
	_, err := c.cfcurl(cli, uri, "-X", "PUT", "-H", "Content-Type: application/json", "-d", fmt.Sprintf(`{"username": "%s"}`, username))
	if nil != err {
		return err
	}

	return nil
}

func (c *Client) SetOrgRole(cli plugin.CliConnection, orgGuid string, username string, orgRole string) error {
	var role string

	switch orgRole {
	case OrgManager:
		role = "managers"
	case OrgBillingManager:
		role = "billing_managers"
	case OrgAuditor:
		role = "auditors"
	}

	uri := fmt.Sprintf("/v2/organizations/%s/%s", orgGuid, role)
	_, err := c.cfcurl(cli, uri, "-X", "PUT", "-H", "Content-Type: application/json", "-d", fmt.Sprintf(`{"username": "%s"}`, username))
	if nil != err {
		return err
	}

	return nil
}

func (c *Client) GetOrgByName(cli plugin.CliConnection, orgName string) (*OrgResource, error) {
	if org, ok := c.orgsByName[orgName]; ok {
		return org, nil
	}

	orgsResponse, err := c.FindOrg(cli, orgName)
	if err != nil {
		return nil, fmt.Errorf("Failed to find org by name '%s'; %v", orgName, err)
	}

	if len(orgsResponse.Resources) == 0 {
		return nil, fmt.Errorf("Org '%s' does not exist", orgName)
	}
	if len(orgsResponse.Resources) != 1 {
		return nil, fmt.Errorf("Found more than one org name matching '%s'", orgName)
	}

	c.orgsByName[orgName] = orgsResponse.Resources[0]
	return orgsResponse.Resources[0], nil
}

func (c *Client) FindOrg(cli plugin.CliConnection, orgName string) (OrgsResponse, error) {
	var orgsResponse OrgsResponse

	data, err := c.cfcurl(cli, fmt.Sprintf("/v2/organizations?q=name:%s", orgName))

	if nil != err {
		return orgsResponse, err
	}

	err = json.Unmarshal(data, &orgsResponse)

	if nil != err {
		fmt.Println("Failed to parse json: ", err.Error())
		return orgsResponse, err
	}

	return orgsResponse, err
}

func (c *Client) SetSpaceRoles(cli plugin.CliConnection, username string, spaceRoles []*SpaceRole) error {
	for _, spaceRole := range spaceRoles {
		orgResource, err := c.GetOrgByName(cli, spaceRole.OrgName)
		if err != nil {
			log.Println(err.Error())
			continue
		}

		orgGuid := orgResource.Metadata.GUID
		spacesResource, err := c.GetSpaceByName(cli, orgGuid, spaceRole.SpaceName)
		if err != nil {
			log.Println(err.Error())
			continue
		}

		spaceGuid := spacesResource.Metadata.GUID
		if err := c.SetSpaceRole(cli, spaceGuid, spaceRole.RoleName, username); err != nil {
			fmt.Printf("Failed to set space role %s for user %s on space %s and org %s; %v\n", spaceRole.RoleName, username, spaceRole.RoleName, spaceRole.OrgName, err)
		}
	}

	return nil
}

func (c *Client) SetSpaceRole(cli plugin.CliConnection, spaceGuid string, spaceRole string, username string) error {
	var role string

	switch spaceRole {
	case SpaceDeveloper:
		role = "developers"
	case SpaceAuditor:
		role = "auditors"
	case SpaceManager:
		role = "managers"
	}

	uri := fmt.Sprintf("/v2/spaces/%s/%s", spaceGuid, role)
	_, err := c.cfcurl(cli, uri, "-X", "PUT", "-d", fmt.Sprintf(`{"username": "%s"}`, username))
	if nil != err {
		return err
	}

	return nil
}

func (c *Client) GetSpaceByName(cli plugin.CliConnection, orgGuid string, spaceName string) (*SpaceResource, error) {
	spaceKey := fmt.Sprintf("%s-%s", orgGuid, spaceName)
	if spaceResource, ok := c.spacesByName[spaceKey]; ok {
		return spaceResource, nil
	}

	spacesResponse, err := c.FindSpace(cli, orgGuid, spaceName)
	if err != nil {
		return nil, fmt.Errorf("Failed to find space %s in org %s; %v", spaceName, orgGuid, err)
	}
	if len(spacesResponse.Resources) == 0 {
		return nil, fmt.Errorf("Space '%s' in Org %s does not exist", spaceName, orgGuid)
	}
	if len(spacesResponse.Resources) != 1 {
		return nil, fmt.Errorf("Found more than one space matching name '%s'", spaceName)
	}

	c.spacesByName[spaceKey] = spacesResponse.Resources[0]
	return spacesResponse.Resources[0], nil
}

func (c *Client) FindSpace(cli plugin.CliConnection, orgGuid string, spaceName string) (SpacesResponse, error) {
	var spacesResponse SpacesResponse

	data, err := c.cfcurl(cli, fmt.Sprintf("/v2/spaces?q=name:%s&q=organization_guid:%s", spaceName, orgGuid))

	if nil != err {
		return spacesResponse, err
	}

	err = json.Unmarshal(data, &spacesResponse)

	if nil != err {
		fmt.Println("Failed to parse json: ", err.Error())
		return spacesResponse, err
	}

	return spacesResponse, err
}

func (c *Client) cfcurl(cli plugin.CliConnection, cliCommandArgs ...string) (data []byte, err error) {
	cliCommandArgs = append([]string{"curl"}, cliCommandArgs...)
	output, err := cli.CliCommandWithoutTerminalOutput(cliCommandArgs...)
	if nil != err {
		return nil, err
	}

	if nil == output || 0 == len(output) {
		return nil, errors.New("CF API returned no output")
	}

	response := strings.Join(output, " ")
	if 0 == len(response) || "" == response {
		return nil, errors.New("Failed to join output")
	}

	return []byte(response), err
}
