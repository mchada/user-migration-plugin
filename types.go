package main

import "fmt"

type PagedResponse struct {
	TotalResults int    `json:"total_results"`
	TotalPages   int    `json:"total_pages"`
	PrevUrl      string `json:"prev_url"`
	NextUrl      string `json:"next_url"`
}

type UsersResponse struct {
	*PagedResponse
	Resources []*UserResource `json:"resources"`
}

type ResourceMetadata struct {
	URL  string `json:"url"`
	GUID string `json:"guid"`
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
	Active   bool   `json:"admin"`
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
	Org      *OrgResource
	RoleName string
}

func (o OrgRole) String() string {
	return fmt.Sprintf("%s %s", o.getOrgName(), o.RoleName)
}

func (o *OrgRole) getOrgName() string {
	return o.Org.Entity.Name
}

type Space struct {
	Name string `json:"name"`
}

type SpaceRole struct {
	Org      *OrgResource
	Space    *SpaceResource
	RoleName string
}

func (s *SpaceRole) getOrgName() string {
	return s.Org.Entity.Name
}

func (s *SpaceRole) getSpaceName() string {
	return s.Space.Entity.Name
}

func (s SpaceRole) String() string {
	return fmt.Sprintf("%s %s %s", s.getOrgName(), s.getSpaceName(), s.RoleName)
}

func (u *UserSummary) getOrgRoles() []*OrgRole {
	orgRoles := make([]*OrgRole, 0)
	orgRoles = append(orgRoles, getOrgRoles(u.ManagedOrganizations, "OrgManager")...)
	orgRoles = append(orgRoles, getOrgRoles(u.BillingManagedOrganizations, "BillingManager")...)
	orgRoles = append(orgRoles, getOrgRoles(u.AuditedManagedOrganizations, "OrgAuditor")...)

	return orgRoles
}

func getOrgRoles(orgs []*OrgResource, orgRoleName string) []*OrgRole {
	orgRoles := make([]*OrgRole, 0)
	for _, orgResource := range orgs {
		orgRoles = append(orgRoles, &OrgRole{orgResource, orgRoleName})
	}

	return orgRoles
}

func (u *UserSummary) getSpaceRoles() []*SpaceRole {
	spaceRoles := make([]*SpaceRole, 0)
	spaceRoles = append(spaceRoles, getSpaceRoles(u, u.Spaces, "SpaceDeveloper")...)
	spaceRoles = append(spaceRoles, getSpaceRoles(u, u.ManagedSpaces, "SpaceManager")...)
	spaceRoles = append(spaceRoles, getSpaceRoles(u, u.AuditedSpaces, "SpaceAuditor")...)

	return spaceRoles
}

func getSpaceRoles(u *UserSummary, spaces []*SpaceResource, spaceRoleName string) []*SpaceRole {
	spaceRoles := make([]*SpaceRole, 0)

	for _, spaceResource := range spaces {
		spaceRoles = append(spaceRoles, &SpaceRole{u.getOrg(spaceResource), spaceResource, spaceRoleName})
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
