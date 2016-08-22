package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/cloudfoundry/cli/plugin"
	"github.com/kelseyhightower/envconfig"
	"github.com/pivotalservices/user-migration-plugin/uaa"
)

const (
	pluginName string = "user-migration"
)

type userMigration struct {
	UID        string
	Username   string
	ExternalID string
	Emails     []uaa.UserEmail
	OrgRoles   []*OrgRole
	SpaceRoles []*SpaceRole
}

//UserMigrationCmd the plugin
type UserMigrationCmd struct {
}

//GetMetadata returns metatada
func (cmd *UserMigrationCmd) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: pluginName,
		Version: plugin.VersionType{
			Major: 1,
			Minor: 0,
			Build: 0,
		},
		Commands: []plugin.Command{
			{
				Name:     pluginName,
				HelpText: "Pulls all Org and Space users from a CF deployment and migrates them all to another CF deployment",
				UsageDetails: plugin.Usage{
					Usage: "cf user-migration report\n\n",
				},
			},
		},
	}
}

func (cmd *UserMigrationCmd) Run(cli plugin.CliConnection, args []string) {
	cmd.UserMigrationCommand(cli, args)
}

func main() {
	plugin.Start(new(UserMigrationCmd))
}

func (cmd *UserMigrationCmd) UserMigrationCommand(cli plugin.CliConnection, args []string) {
	if nil == cli {
		fmt.Println("ERROR: CLI Connection is nil!")
		os.Exit(1)
	}

	if isLoggedIn, err := cli.IsLoggedIn(); err == nil && isLoggedIn != true {
		fmt.Println("You are not logged in. Please login using 'cf login' and try again")
		os.Exit(1)
	}

	if args[0] == pluginName && args[1] == "report" {
		cmd.printUserReport(cli)
	}

}

func (cmd *UserMigrationCmd) printUserReport(cli plugin.CliConnection) {
	apiEndpoint, err := cli.ApiEndpoint()
	if err != nil {
		fmt.Println("Failed to get api endpoint from plugin.CliConnection: ", err.Error())
	}

	uaaEndpoint := strings.Replace(apiEndpoint, "api", "uaa", 1)
	fmt.Println("derived uaa endpoint: ", uaaEndpoint)

	uaaConnInfo := uaa.ConnectionInfo{ServerURL: uaaEndpoint}
	err = envconfig.Process("uaa", &uaaConnInfo)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	uaac, err := uaaConnInfo.Connect()
	if err != nil {
		fmt.Println("Failed to connect to UAA: ", err.Error())
		os.Exit(1)
	}

	uaaUsers, err := uaac.ListUsers()
	if err != nil {
		fmt.Println("Failed to list users from UAA: ", err.Error())
		return
	}

	userMigrations := make([]*userMigration, 0)
	usersResponse, _ := cmd.getUsers(cli)
	for _, userResource := range usersResponse.Resources {
		if len(userResource.Entity.Username) == 0 {
			fmt.Printf("User with GUID %s has no username in CC :(\n", userResource.Metadata.GUID)
			continue
		}

		userMigration := new(userMigration)
		userMigration.Username = userResource.Entity.Username

		uaaUser := findUaaUser(userResource, &uaaUsers)
		if uaaUser == nil {
			fmt.Printf("UAA User not found for CC user with username%s\n", userResource.Entity.Username)
			continue
		}

		userMigration.ExternalID = uaaUser.ExternalID
		userMigration.Emails = uaaUser.Emails

		userSummary, _ := cmd.getUserSummary(cli, userResource)
		userMigration.OrgRoles = userSummary.Entity.getOrgRoles()
		userMigration.SpaceRoles = userSummary.Entity.getSpaceRoles()

		userMigrations = append(userMigrations, userMigration)
	}

	b, err := json.MarshalIndent(userMigrations, "", "  ")
	if err != nil {
		fmt.Println("error marshalling user migrations to json:", err)
	}
	os.Stdout.Write(b)

}

func findUaaUser(userResource *UserResource, uaaUsers *uaa.Users) *uaa.User {
	for _, uaaUser := range uaaUsers.Users {
		if uaaUser.Username == userResource.Entity.Username {
			return &uaaUser
		}
	}

	return nil
}

func (cmd *UserMigrationCmd) getUsers(cli plugin.CliConnection) (UsersResponse, error) {
	var usersResponse UsersResponse

	data, err := cmd.cfcurl(cli, "/v2/users?results-per-page=100")

	if nil != err {
		return usersResponse, err
	}

	err = json.Unmarshal(data, &usersResponse)

	if nil != err {
		fmt.Println("Failed to parse json: ", err.Error())
		return usersResponse, err
	}

	return usersResponse, err
}

func (cmd *UserMigrationCmd) getUserSummary(cli plugin.CliConnection, userResource *UserResource) (UserSummaryResource, error) {
	var userSummaryResource UserSummaryResource

	data, err := cmd.cfcurl(cli, fmt.Sprintf("/v2/users/%s/summary", userResource.Metadata.GUID))

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

func (cmd *UserMigrationCmd) cfcurl(cli plugin.CliConnection, cliCommandArgs ...string) (data []byte, err error) {
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
