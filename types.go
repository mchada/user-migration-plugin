package main

import (
	"github.com/dave-malone/cfclient"
	"github.com/pivotalservices/go-uaac/users"
)

type userMigration struct {
	Username   string
	ExternalID string
	Emails     []users.UserEmail
	OrgRoles   []*cf.OrgRole
	SpaceRoles []*cf.SpaceRole
}

type userExport struct {
	CfApiUrl       string
	UserMigrations []*userMigration
}
