package main

import (
	"github.com/dave-malone/cfclient"
	"github.com/dave-malone/go-uaac"
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
