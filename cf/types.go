package cf

import "fmt"

const (
	OrgManager        string = "OrgManager"
	OrgBillingManager string = "BillingManager"
	OrgAuditor        string = "OrgAuditor"
	SpaceDeveloper    string = "SpaceDeveloper"
	SpaceManager      string = "SpaceManager"
	SpaceAuditor      string = "SpaceAuditor"
)

type PagedResponse struct {
	TotalResults int    `json:"total_results"`
	TotalPages   int    `json:"total_pages"`
	PrevUrl      string `json:"prev_url"`
	NextUrl      string `json:"next_url"`
}

type ResourceMetadata struct {
	URL  string `json:"url"`
	GUID string `json:"guid"`
}

type UsersResponse struct {
	*PagedResponse
	Resources []*UserResource `json:"resources"`
}

type OrgsResponse struct {
	*PagedResponse
	Resources []*OrgResource `json:"resources"`
}

type SpacesResponse struct {
	*PagedResponse
	Resources []*SpaceResource `json:"resources"`
}

type UserResource struct {
	Entity   User             `json:"entity"`
	Metadata ResourceMetadata `json:"metadata"`
}

type UserSummaryResource struct {
	Entity   UserSummary      `json:"entity"`
	Metadata ResourceMetadata `json:"metadata"`
}

type OrgResource struct {
	Entity   Org              `json:"entity"`
	Metadata ResourceMetadata `json:"metadata"`
}

type SpaceResource struct {
	Entity   Space            `json:"entity"`
	Metadata ResourceMetadata `json:"metadata"`
}

type User struct {
	Admin    bool   `json:"admin"`
	Active   bool   `json:"active"`
	Username string `json:"username"`
}

type UserSummary struct {
	Organizations               []*OrgResource   `json:"organizations"`
	ManagedOrganizations        []*OrgResource   `json:"managed_organizations"`
	BillingManagedOrganizations []*OrgResource   `json:"billing_managed_organizations"`
	AuditedManagedOrganizations []*OrgResource   `json:"audited_managed_organizations"`
	Spaces                      []*SpaceResource `json:"spaces"`
	ManagedSpaces               []*SpaceResource `json:"managed_spaces"`
	AuditedSpaces               []*SpaceResource `json:"audited_spaces"`
}

//Organization representation
type Org struct {
	Name   string           `json:"name"`
	Spaces []*SpaceResource `json:"spaces"`
}

type OrgRole struct {
	Org      *OrgResource `json:"-"`
	OrgName  string
	RoleName string
}

func (o OrgRole) String() string {
	return fmt.Sprintf("%s %s", o.OrgName, o.RoleName)
}

type Space struct {
	Name string `json:"name"`
}

type SpaceRole struct {
	Org       *OrgResource   `json:"-"`
	Space     *SpaceResource `json:"-"`
	OrgName   string
	SpaceName string
	RoleName  string
}

func (s SpaceRole) String() string {
	return fmt.Sprintf("%s %s %s", s.OrgName, s.SpaceName, s.RoleName)
}

func (u *UserSummary) GetOrgRoles() []*OrgRole {
	orgRoles := make([]*OrgRole, 0)
	orgRoles = append(orgRoles, getOrgRoles(u.ManagedOrganizations, OrgManager)...)
	orgRoles = append(orgRoles, getOrgRoles(u.BillingManagedOrganizations, OrgBillingManager)...)
	orgRoles = append(orgRoles, getOrgRoles(u.AuditedManagedOrganizations, OrgAuditor)...)

	return orgRoles
}

func getOrgRoles(orgs []*OrgResource, orgRoleName string) []*OrgRole {
	orgRoles := make([]*OrgRole, 0)
	for _, orgResource := range orgs {
		orgRoles = append(orgRoles, &OrgRole{Org: orgResource,
			RoleName: orgRoleName,
			OrgName:  orgResource.Entity.Name,
		})
	}

	return orgRoles
}

func (u *UserSummary) GetSpaceRoles() []*SpaceRole {
	spaceRoles := make([]*SpaceRole, 0)
	spaceRoles = append(spaceRoles, getSpaceRoles(u, u.Spaces, SpaceDeveloper)...)
	spaceRoles = append(spaceRoles, getSpaceRoles(u, u.ManagedSpaces, SpaceManager)...)
	spaceRoles = append(spaceRoles, getSpaceRoles(u, u.AuditedSpaces, SpaceAuditor)...)

	return spaceRoles
}

func getSpaceRoles(u *UserSummary, spaces []*SpaceResource, spaceRoleName string) []*SpaceRole {
	spaceRoles := make([]*SpaceRole, 0)

	for _, spaceResource := range spaces {
		org := u.getOrg(spaceResource)
		spaceRoles = append(spaceRoles, &SpaceRole{Org: u.getOrg(spaceResource),
			Space:     spaceResource,
			RoleName:  spaceRoleName,
			OrgName:   org.Entity.Name,
			SpaceName: spaceResource.Entity.Name,
		})
	}

	return spaceRoles
}

func (u *UserSummary) getOrg(space *SpaceResource) *OrgResource {
	for _, org := range u.Organizations {
		for _, orgSpace := range org.Entity.Spaces {
			if orgSpace.Metadata.GUID == space.Metadata.GUID {
				return org
			}
		}
	}

	return nil
}
