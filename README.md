user-migration plugin
======


Retrieves the list of Users from the Cloud Controller API and the UAA API and outputs a json file that can then be edited by hand, if needed, to then import into another instance of CF


# Install the plugin:

`cf install-plugin user-migration-plugin`

# Run it

1. target your old environment:

```
export UAA_CLIENTID=admin
export UAA_CLIENTSECRET=old-env-admin-client-secret
cf login -a https://api.oldsystemdomain --skip-ssl-validation
cf user-migration export user-migration.json
```

2. target your new environment:

```
export UAA_CLIENTID=admin
export UAA_CLIENTSECRET=new-env-admin-client-secret
cf login -a https://api.newsystemdomain --skip-ssl-validation
cf user-migration import user-migration.json
```


## Build the plugin

Build it for Linux:

`GOOS=linux GOARCH=amd64 go build`
