module github.com/Confialink/wallet-accounts

go 1.13

replace (
	github.com/Confialink/wallet-accounts/rpc/accounts => ./rpc/accounts
	github.com/Confialink/wallet-accounts/rpc/limit => ./rpc/limit
)

require (
	github.com/Confialink/wallet-accounts/rpc/accounts v0.0.0-00010101000000-000000000000
	github.com/Confialink/wallet-accounts/rpc/limit v0.0.0-00010101000000-000000000000
	github.com/Confialink/wallet-currencies/rpc/currencies v0.0.0-20210218071616-68af0b7e4375
	github.com/Confialink/wallet-logs/rpc/logs v0.0.0-20210218064020-81b818342efd
	github.com/Confialink/wallet-notifications/rpc/proto/notifications v0.0.0-20210218064438-818cea3b20db
	github.com/Confialink/wallet-permissions/rpc/permissions v0.0.0-20210218064621-7b7ddad868c8
	github.com/Confialink/wallet-pkg-acl v0.0.0-20210218070839-a03813da4b89
	github.com/Confialink/wallet-pkg-custom_form v0.0.0-20210217125758-eb25fcc8b328
	github.com/Confialink/wallet-pkg-discovery/v2 v2.0.0-20210217105157-30e31661c1d1
	github.com/Confialink/wallet-pkg-env_config v0.0.0-20210217112253-9483d21626ce
	github.com/Confialink/wallet-pkg-env_mods v0.0.0-20210217112432-4bda6de1ee2c
	github.com/Confialink/wallet-pkg-errors v1.0.2
	github.com/Confialink/wallet-pkg-list_params v0.0.0-20210217104359-69dfc53fe9ee
	github.com/Confialink/wallet-pkg-model_serializer v0.0.0-20210217111055-c5e1cb1a75c7
	github.com/Confialink/wallet-pkg-response v0.0.0-20210218060251-4315861fed87
	github.com/Confialink/wallet-pkg-service_names v0.0.0-20210217112604-179d69540dea
	github.com/Confialink/wallet-pkg-types v0.0.0-20210217112028-920c305b25ec
	github.com/Confialink/wallet-pkg-utils v0.0.0-20210217112822-e79f6d74cdc1
	github.com/Confialink/wallet-settings/rpc/proto/settings v0.0.0-20210218070334-b4153fc126a0
	github.com/Confialink/wallet-users/rpc/proto/users v0.0.0-20210218071418-0600c0533fb2
	github.com/DATA-DOG/go-sqlmock v1.5.0
	github.com/SebastiaanKlippert/go-wkhtmltopdf v1.6.1
	github.com/fatih/structs v1.1.0
	github.com/gin-contrib/cors v1.3.1
	github.com/gin-gonic/gin v1.6.3
	github.com/go-playground/validator/v10 v10.3.0
	github.com/golang/mock v1.4.4
	github.com/google/go-cmp v0.5.2 // indirect
	github.com/iancoleman/strcase v0.0.0-20191112232945-16388991a334
	github.com/inconshreveable/log15 v0.0.0-20200109203555-b30bc20e4fd1
	github.com/jinzhu/gorm v1.9.15
	github.com/jinzhu/now v1.1.1
	github.com/json-iterator/go v1.1.10
	github.com/kildevaeld/go-acl v0.0.0-20171228130000-7799b11f4759
	github.com/kr/text v0.2.0 // indirect
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/olebedev/emitter v0.0.0-20190110104742-e8d1457e6aee
	github.com/onsi/ginkgo v1.14.0
	github.com/onsi/gomega v1.10.1
	github.com/pkg/errors v0.9.1
	github.com/robfig/cron v1.2.0
	github.com/shopspring/decimal v1.2.0
	github.com/stretchr/objx v0.1.1 // indirect
	github.com/stretchr/testify v1.6.1
	go.uber.org/dig v1.10.0
	golang.org/x/crypto v0.0.0-20200709230013-948cd5f35899
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b // indirect
	golang.org/x/sys v0.0.0-20200905004654-be1d3432aa8f // indirect
	golang.org/x/text v0.3.3 // indirect
	golang.org/x/tools v0.0.0-20200904185747-39188db58858 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/yaml.v3 v3.0.0-20200605160147-a5ece683394c // indirect
)
