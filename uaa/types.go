package uaa

type ServerInfo struct {
	App            map[string]string
	Links          map[string]string
	ZoneName       string `json:"zone_name"`
	EntityID       string
	CommitID       string `json:"commit_id"`
	IDPDefinitions map[string]string
	Prompts        map[string][]string
	Timestamp      string
}

func (s ServerInfo) Version() string {
	return s.App["version"]
}

type AccessToken struct {
	Token     string `json:"access_token"`
	Type      string `json:"token_type"`
	ExpiresIn int    `json:"expires_in"`
	Scope     string
	JTI       string
}

type OauthClients struct {
	Clients      []OauthClient `json:"resources"`
	StartIndex   int
	ItemsPerPage int
	TotalResults int
	Schemas      []string
}

type OauthClient struct {
	ID   string `json:"client_id"`
	Name string
	//AutoApprove          bool - TODO - this field can be a bool or an array??
	Action                 string
	Scope                  []string
	ResourceIDs            []string `json:"resource_ids"`
	Authorities            []string
	AuthorizedGrantTypes   []string `json:"authorized_grant_types"`
	LastModified           int
	RedirectURI            []string `json:"redirect_uri"`
	SignupRedirectURL      string   `json:"signup_redirect_url"`
	ChangeEmailRedirectURL string   `json:"change_email_redirect_url"`
}

type IdentityZone struct {
	ID           string
	Subdomain    string
	Name         string
	Version      int
	Description  string
	Created      int
	LastModified int `json:"last_modified"`
}

type Users struct {
	Users        []User `json:"resources"`
	StartIndex   int
	ItemsPerPage int
	TotalResults int
	Schemas      []string
}

type User struct {
	GUID                 string `json:"id"`
	ExternalID           string
	Username             string
	Name                 Name
	Groups               []Group
	Emails               []UserEmail
	Active               bool
	Verified             bool
	Origin               string
	ZoneID               string
	PasswordLastModified string
	Schemas              []string
}

type Name struct {
	GivenName  string
	FamilyName string
}

type UserEmail struct {
	Value   string
	Primary bool
}

type Approval struct {
	UserID        string
	ClientID      string
	Scope         string
	Status        string
	LastUpdatedAt string
	ExpiresAt     string
}

type Group struct {
	Value   string
	Display string
	Type    string
}
