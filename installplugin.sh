rm user-migration-plugin
cf uninstall-plugin user-migration
go build
cf install-plugin user-migration-plugin
