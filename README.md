user-migration plugin
======


Retrieves the list of Users from the Cloud Controller API and the UAA API and outputs a yml file which can be used with the [UAA Ldap Import](https://github.com/pivotalservices/uaaldapimport) tool.


Since this plugin requires access to UAA, the following environment variables are required:

```
export UAA_SERVERURL=https://uaa.local.pcfdev.io
export UAA_CLIENTID=admin
export UAA_CLIENTSECRET=admin-client-secret
```
