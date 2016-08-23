package cf

import "github.com/pivotalservices/user-migration-plugin/uaa"

type Client interface {
	GetUsers() (UsersResponse, error)
	CreateUser(guid uaa.UserGuid) error
	GetUserSummary(userResource *UserResource) (UserSummaryResource, error)
	SetOrgRoles(guid uaa.UserGuid, orgRoles []*OrgRole) error
	AssociateUserWithOrg(orgGuid string, guid uaa.UserGuid) error
	SetOrgRole(orgGuid string, guid uaa.UserGuid, orgRole string) error
	GetOrgByName(orgName string) (*OrgResource, error)
	FindOrg(orgName string) (OrgsResponse, error)
	SetSpaceRoles(guid uaa.UserGuid, spaceRoles []*SpaceRole) error
	SetSpaceRole(spaceGuid string, spaceRole string, guid uaa.UserGuid) error
	GetSpaceByName(orgGuid string, spaceName string) (*SpaceResource, error)
	FindSpace(orgGuid string, spaceName string) (SpacesResponse, error)
}
