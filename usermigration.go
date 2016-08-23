package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/cloudfoundry/cli/plugin"
	"github.com/dave-malone/cfclient"
	"github.com/dave-malone/go-uaac"
	"github.com/kelseyhightower/envconfig"
)

const (
	pluginName string = "user-migration"
)

//UserMigrationCmd the plugin
type UserMigrationCmd struct {
}

func main() {
	plugin.Start(new(UserMigrationCmd))
}

func (cmd *UserMigrationCmd) Run(cli plugin.CliConnection, args []string) {
	cmd.UserMigrationCommand(cli, args)
}

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
				HelpText: "Creates an Export of all Users and their Orgs and Spaces by dumping to a specified. Then, you may import the file into a different deployment. Be sure to change your CF API Target and your UAA_CLIENTID and UAA_CLIENTSECRET between export and import!",
				UsageDetails: plugin.Usage{
					Usage: "cf user-migration export FILE_NAME\n\n   cf user-migration import FILE_NAME",
				},
			},
		},
	}
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

	log.SetOutput(os.Stdout)

	if args[0] == pluginName && args[1] == "export" {
		cmd.exportUsers(cli, args[2])
	} else if args[0] == pluginName && args[1] == "import" {
		cmd.importUsers(cli, args[2])
	}
}

func (cmd *UserMigrationCmd) exportUsers(cli plugin.CliConnection, exportFileName string) {
	uaac := getUaac(cli)
	cfclient := cf.NewCliClient(cli)

	userExport := new(userExport)
	userExport.CfApiUrl = getApiEndpoint(cli)

	uaaUsers, err := uaac.ListUsers()
	if err != nil {
		log.Fatalf("Failed to list users from UAA: %v", err)
	}

	userMigrations := make([]*userMigration, 0)
	usersResponse, _ := cfclient.GetUsers()
	for _, userResource := range usersResponse.Resources {
		if len(userResource.Entity.Username) == 0 {
			fmt.Printf("User with GUID %s has no username in CC :(\n", userResource.Metadata.GUID)
			continue
		}

		userMigration := new(userMigration)
		userMigration.Username = userResource.Entity.Username

		var userSummary cf.UserSummaryResource
		userSummary, err = cfclient.GetUserSummary(userResource)
		if err != nil {
			fmt.Printf("Failed to get user summary for user %s; %v\n", userResource.Entity.Username, err)
			continue
		}
		userMigration.OrgRoles = userSummary.Entity.GetOrgRoles()
		userMigration.SpaceRoles = userSummary.Entity.GetSpaceRoles()

		uaaUser := findUaaUser(userResource, &uaaUsers)
		if uaaUser == nil {
			fmt.Printf("UAA User not found for CC user with username%s\n", userResource.Entity.Username)
			continue
		}

		if len(uaaUser.ExternalID) == 0 {
			fmt.Printf("User with GUID %s does not have an ExternalID in UAA\nuaa user:%v\n", userResource.Metadata.GUID, uaaUser)
			continue
		}

		userMigration.ExternalID = uaaUser.ExternalID
		userMigration.Emails = uaaUser.Emails

		userMigrations = append(userMigrations, userMigration)
	}

	userExport.UserMigrations = userMigrations
	b, err := json.MarshalIndent(userExport, "", "  ")
	if err != nil {
		fmt.Println("error marshalling user migrations to json:", err)
	}

	if err = ioutil.WriteFile(exportFileName, b, 0755); err != nil {
		fmt.Printf("Failed to write output to file %s; err: %v", exportFileName, err)
	}
}

func (cmd *UserMigrationCmd) importUsers(cli plugin.CliConnection, exportFileName string) {
	fileData, err := ioutil.ReadFile(exportFileName)

	if err != nil {
		log.Fatalf("Failed to read %s. Error: %v", exportFileName, err)
	}

	var export userExport

	err = json.Unmarshal(fileData, &export)
	if err != nil {
		log.Fatalf("Failed to unmarshall json to userExport: %v", err)
	}

	if getApiEndpoint(cli) == export.CfApiUrl {
		log.Fatalf("You are currently targeting the cf deployment which users were exported from. Please target the new environment using 'cf api' and run this command again")
	}

	uaac := getUaac(cli)
	cfclient := cf.NewCliClient(cli)

	fmt.Printf("importing %d users\n", len(export.UserMigrations))
	for i, userMigration := range export.UserMigrations {
		uaaUser := &uaa.User{
			Username:   userMigration.Username,
			ExternalID: userMigration.ExternalID,
			Emails:     userMigration.Emails,
			Origin:     "ldap",
		}

		userGuid, err := uaac.CreateUser(uaaUser)
		if err != nil {
			fmt.Printf("%d Failed to create uaa user: %v\n", i, err)
			continue
		}

		if len(userGuid) == 0 {
			fmt.Printf("%d uaa user guid not found for username %s\n", i, userMigration.Username)
			continue
		}

		if err := cfclient.CreateUser(userGuid); err != nil {
			fmt.Printf("%d Failed to create cf user %s: %v\n", i, userMigration.Username, err)
			continue
		}

		if err := cfclient.SetOrgRoles(userGuid, userMigration.OrgRoles); err != nil {
			fmt.Printf("%d Failed to set org roles for cf user %s; %v\n", i, userMigration.Username, err)
		}

		if err := cfclient.SetSpaceRoles(userGuid, userMigration.SpaceRoles); err != nil {
			fmt.Printf("%d Failed to set space roles for cf user %s; %v\n", i, userMigration.Username, err)
		}
		fmt.Printf("%d %s imported\n", i, userMigration.Username)
	}

}

func getApiEndpoint(cli plugin.CliConnection) string {
	apiEndpoint, err := cli.ApiEndpoint()
	if err != nil {
		log.Fatalf("Failed to get api endpoint from plugin.CliConnection: %v", err)
	}
	return apiEndpoint
}

func getUaac(cli plugin.CliConnection) uaa.Client {
	apiEndpoint := getApiEndpoint(cli)
	uaaEndpoint := strings.Replace(apiEndpoint, "api", "uaa", 1)
	fmt.Println("using derived uaa endpoint: ", uaaEndpoint)

	uaaConnInfo := uaa.ConnectionInfo{ServerURL: uaaEndpoint}

	if err := envconfig.Process("uaa", &uaaConnInfo); err != nil {
		log.Fatalf("Failed to read process required environment variables: %v", err)
	}

	uaac, err := uaaConnInfo.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to UAA: %v", err)
	}

	return uaac
}

func findUaaUser(userResource *cf.UserResource, uaaUsers *uaa.Users) *uaa.User {
	for _, uaaUser := range uaaUsers.Users {
		if uaaUser.Username == userResource.Entity.Username {
			return &uaaUser
		}
	}

	return nil
}
