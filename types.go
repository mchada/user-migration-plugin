package main

import (
	"github.com/pivotalservices/user-migration-plugin/cf"
	"github.com/pivotalservices/user-migration-plugin/uaa"
)

type userMigration struct {
	Username   string
	ExternalID string
	Emails     []uaa.UserEmail
	OrgRoles   []*cf.OrgRole
	SpaceRoles []*cf.SpaceRole
}

type userExport struct {
	CfApiUrl       string
	UserMigrations []*userMigration
}
