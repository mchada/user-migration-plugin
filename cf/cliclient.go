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

type cliClient struct {
	cli              plugin.CliConnection
	orgsByName       map[string]*OrgResource
	spacesByName     map[string]*SpaceResource
	assignedUserOrgs []string
}

var assignedUserOrgKey func(guid uaa.UserGuid, org *OrgResource) string = func(guid uaa.UserGuid, org *OrgResource) string {
	return fmt.Sprintf("%s-%s", org.Metadata.GUID, guid)
}

func (c *cliClient) assignUserOrg(guid uaa.UserGuid, org *OrgResource) {
	if c.assignedToOrg(guid, org) {
		return
	}

	c.assignedUserOrgs = append(c.assignedUserOrgs, assignedUserOrgKey(guid, org))
}

func (c *cliClient) assignedToOrg(guid uaa.UserGuid, org *OrgResource) bool {
	key := assignedUserOrgKey(guid, org)
	for _, assignedUserOrg := range c.assignedUserOrgs {
		if key == assignedUserOrg {
			return true
		}
	}
	return false
}

func NewCliClient(cli plugin.CliConnection) Client {
	return &cliClient{
		cli:              cli,
		orgsByName:       make(map[string]*OrgResource, 0),
		spacesByName:     make(map[string]*SpaceResource, 0),
		assignedUserOrgs: make([]string, 0),
	}
}

func (c *cliClient) GetUsers() (UsersResponse, error) {
	var usersResponse UsersResponse
	data, err := c.cfcurl("/v2/users?results-per-page=100")
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

func (c *cliClient) CreateUser(guid uaa.UserGuid) error {
	payload := fmt.Sprintf(`{"guid": "%s"}`, guid)
	_, err := c.cfcurl("/v2/users", "-X", "POST", "-d", payload)
	if nil != err {
		return err
	}

	return nil
}

func (c *cliClient) GetUserSummary(userResource *UserResource) (UserSummaryResource, error) {
	var userSummaryResource UserSummaryResource

	data, err := c.cfcurl(fmt.Sprintf("/v2/users/%s/summary", userResource.Metadata.GUID))

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

func (c *cliClient) SetOrgRoles(guid uaa.UserGuid, orgRoles []*OrgRole) error {
	for _, orgRole := range orgRoles {
		orgResource, err := c.GetOrgByName(orgRole.OrgName)
		if err != nil {
			log.Printf("%v", err)
			continue
		}

		orgGuid := orgResource.Metadata.GUID

		if c.assignedToOrg(guid, orgResource) != true {
			if err := c.AssociateUserWithOrg(orgGuid, guid); err != nil {
				fmt.Printf("failed to associate user %s with org %s; %v", guid, orgRole.OrgName, err)
			}
			c.assignUserOrg(guid, orgResource)
		}

		if err := c.SetOrgRole(orgGuid, guid, orgRole.RoleName); err != nil {
			fmt.Printf("failed to set org role; user %s - org %s - org role: %s; %v", guid, orgRole.OrgName, orgRole.RoleName, err)
		}
	}

	return nil
}

func (c *cliClient) AssociateUserWithOrg(orgGuid string, guid uaa.UserGuid) error {
	uri := fmt.Sprintf("/v2/organizations/%s/users/%s", orgGuid, guid)
	data, err := c.cfcurl(uri, "-X", "PUT")
	if nil != err {
		return err
	}

	var response map[string]interface{}
	err = json.Unmarshal(data, &response)
	if err != nil {
		return fmt.Errorf("Failed to process PUT:%s response: %v", uri, err)
	}

	if _, ok := response["error_code"]; ok {
		return fmt.Errorf("Error setting space role %v\n", response)
	}

	return nil
}

func (c *cliClient) SetOrgRole(orgGuid string, guid uaa.UserGuid, orgRole string) error {
	var role string

	switch orgRole {
	case OrgManager:
		role = "managers"
	case OrgBillingManager:
		role = "billing_managers"
	case OrgAuditor:
		role = "auditors"
	}

	uri := fmt.Sprintf("/v2/organizations/%s/%s/%s", orgGuid, role, guid)
	data, err := c.cfcurl(uri, "-X", "PUT")
	if nil != err {
		return err
	}

	var setOrgRoleResponse map[string]interface{}
	err = json.Unmarshal(data, &setOrgRoleResponse)
	if err != nil {
		return fmt.Errorf("Failed to process PUT:%s response: %v", uri, err)
	}

	if _, ok := setOrgRoleResponse["error_code"]; ok {
		return fmt.Errorf("Error setting space role %v\n", setOrgRoleResponse)
	}

	return nil
}

func (c *cliClient) GetOrgByName(orgName string) (*OrgResource, error) {
	if org, ok := c.orgsByName[orgName]; ok {
		return org, nil
	}

	orgsResponse, err := c.FindOrg(orgName)
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

func (c *cliClient) FindOrg(orgName string) (OrgsResponse, error) {
	var orgsResponse OrgsResponse

	data, err := c.cfcurl(fmt.Sprintf("/v2/organizations?q=name:%s", orgName))

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

func (c *cliClient) SetSpaceRoles(guid uaa.UserGuid, spaceRoles []*SpaceRole) error {
	for _, spaceRole := range spaceRoles {
		orgResource, err := c.GetOrgByName(spaceRole.OrgName)
		if err != nil {
			log.Println(err.Error())
			continue
		}

		orgGuid := orgResource.Metadata.GUID
		if c.assignedToOrg(guid, orgResource) != true {
			if err = c.AssociateUserWithOrg(orgGuid, guid); err != nil {
				fmt.Printf("failed to associate user %s with org %s; %v", guid, spaceRole.OrgName, err)
			}
			c.assignUserOrg(guid, orgResource)
		}

		spacesResource, err := c.GetSpaceByName(orgGuid, spaceRole.SpaceName)
		if err != nil {
			log.Println(err.Error())
			continue
		}

		spaceGuid := spacesResource.Metadata.GUID
		if err := c.SetSpaceRole(spaceGuid, spaceRole.RoleName, guid); err != nil {
			fmt.Printf("Failed to set space role %s for user %s on space %s and org %s; %v\n", spaceRole.RoleName, guid, spaceRole.RoleName, spaceRole.OrgName, err)
		}
	}

	return nil
}

func (c *cliClient) SetSpaceRole(spaceGuid string, spaceRole string, guid uaa.UserGuid) error {
	var role string

	switch spaceRole {
	case SpaceDeveloper:
		role = "developers"
	case SpaceAuditor:
		role = "auditors"
	case SpaceManager:
		role = "managers"
	}

	uri := fmt.Sprintf("/v2/spaces/%s/%s/%s", spaceGuid, role, guid)
	data, err := c.cfcurl(uri, "-X", "PUT")
	if nil != err {
		return err
	}

	var setSpaceRoleResponse map[string]interface{}
	err = json.Unmarshal(data, &setSpaceRoleResponse)
	if err != nil {
		return fmt.Errorf("Failed to process PUT:%s response: %v", uri, err)
	}

	if _, ok := setSpaceRoleResponse["error_code"]; ok {
		return fmt.Errorf("Error setting space role %v\n", setSpaceRoleResponse)
	}

	return nil
}

func (c *cliClient) GetSpaceByName(orgGuid string, spaceName string) (*SpaceResource, error) {
	spaceKey := fmt.Sprintf("%s-%s", orgGuid, spaceName)
	if spaceResource, ok := c.spacesByName[spaceKey]; ok {
		return spaceResource, nil
	}

	spacesResponse, err := c.FindSpace(orgGuid, spaceName)
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

func (c *cliClient) FindSpace(orgGuid string, spaceName string) (SpacesResponse, error) {
	var spacesResponse SpacesResponse

	data, err := c.cfcurl(fmt.Sprintf("/v2/spaces?q=name:%s&q=organization_guid:%s", spaceName, orgGuid))

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

func (c *cliClient) cfcurl(cliCommandArgs ...string) (data []byte, err error) {
	cliCommandArgs = append([]string{"curl"}, cliCommandArgs...)
	output, err := c.cli.CliCommandWithoutTerminalOutput(cliCommandArgs...)
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
