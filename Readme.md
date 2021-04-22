**Dependencies**

 - Golang V1.10
 - Docker
 - MySql

**Build microservice**
Run the following command from microservice directory:

make build

**Available commands**
There is embedded module that can be used in order to execute predefined commands.
`service_accounts -cmd help`

Following environment variables are required for Velmie Wallet Accounts Service:

 - VELMIE_WALLET_ACCOUNTS_DB_PASS=password
 - VELMIE_WALLET_ACCOUNTS_DB_USER=user
 - VELMIE_WALLET_ACCOUNTS_DB_NAME=schema
 - VELMIE_WALLET_ACCOUNTS_DB_PORT=port
 - VELMIE_WALLET_ACCOUNTS_DB_HOST=host
 - VELMIE_WALLET_ACCOUNTS_DB_DRIV=mysql
 - VELMIE_WALLET_ACCOUNTS_DB_IS_DEBUG_MODE=false
 - VELMIE_WALLET_ACCOUNTS_CORS_HEADERS=*
 - VELMIE_WALLET_ACCOUNTS_CORS_ORIGINS=*
 - VELMIE_WALLET_ACCOUNTS_CORS_METHODS=GET,POST,PUT,PATCH,DELETE,OPTIONS
 - VELMIE_WALLET_ACCOUNTS_SCHEDULED_TASKS_SIMULATION_ENABLED=false

 **List of event**
 | Module      | Event Name           | Constant           |  Arguments                | Description                   |
 |-------------|----------------------|--------------------|---------------------------|-------------------------------|
 | account-type| account-type:updated | AccountTypeUpdated | ContextAccountTypeUpdated | When account type is updated  |

**Migrate schema**

[[Golang migration tool]](https://github.com/golang-migrate/migrate)

    migrate -database mysql://username:password@protocol(host)/dbname -path database/migrations

**Push Image to repository**

    sudo docker push repository/name:version

 - **repository** - dockerhub or AWS ECS Registry repository name
 - **name**  - image name
 - **version** - image version

**Deploy containerized  microservice to AWS ECS**

**Build tags**

Some part of functionality is used for simulations. Those packages are marked by tag ```simulations```. You must build the app with option ```-tags=simulations``` to include features to the app build.

**Scheduled tasks simulation**

You must set ```VELMIE_WALLET_ACCOUNTS_SCHEDULED_TASKS_SIMULATION_ENABLED=true``` to enable 'Scheduled tasks simulation' features. Also it will be working only in dev mode by ```ENV```. It adds the route ```/accounts/admin/simulations/scheduled-tasks```  where an admin cat set up dates of simulation and run all scheduled task manually.

**Initial data**  

See [initial data](docs/initial_data.md)
**Forms and dynamic validation**  

To implement dynamic validation there was created formService see ```wallet--accounts/modules/common/service/forms/service/service.go```. 
It allows to register different types of forms for different models.

When You registered your form in the service You can use one to dynamic validation in an action handler:  

```go
package x

import (
	"github.com/Confialink/dominica-backend/packages/custom_form"
	"github.com/gin-gonic/gin"
	"github.com/Confialink/dominica-backend/packages/errors"
	formModelAccount "github.com/Confialink/dominica-backend/service-accounts/modules/common/service/forms/model/account"
	formService "github.com/Confialink/dominica-backend/service-accounts/modules/common/service/forms/service"
	"github.com/Confialink/dominica-backend/service-accounts/modules/common/service/forms"
	"github.com/inconshreveable/log15"
	appValidator "github.com/Confialink/dominica-backend/service-accounts/modules/app/validator"
)

// request for validation
type Request struct {
	TypeId int32
	Number string
	CustomForm *custom_form.Form `json:"-"`
}

// implement custom_form.FormContainer interface
func (r Request) GetCustomForm() *custom_form.Form {
	return r.CustomForm
}

type controller struct{
	modelFormService *formService.ModelFormService
	validator        appValidator.Interface
	logger           log15.Logger
	//...
}

// action handler
func (h *controller) Action(c *gin.Context) {
	req := Request{}
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.AddShouldBindError(c, err)
		return
	}

	modelName := formModelAccount.ModelName
	formType := forms.TypeFormPost
	modelForm, typedErr := h.modelFormService.MakeForm(currentUser, modelName, formType)
	if typedErr != nil {
		h.logger.Error(typedErr.Error(), "modelName", modelName, "formType", formType)
		errors.AddErrors(c, typedErr)
		return
	}

	// it works only one time, after first call of validator.Struct the struct is cached in the v8.validator
	// so we cannot register a new callback every time
	// so we put custom form into the struct
	customValidator := h.modelFormService.MakeValidationCallback()
	h.validator.RegisterStructValidation(customValidator, req)

	req.CustomForm = modelForm
	err := h.validator.Struct(req)

	if err != nil {
		errors.AddShouldBindError(c, err)
		return
	}

	// do something
}

```  

The frontend also can receive rules for any form using the route ```/accounts/private/v1/form/{model}/{type}```. For example ```/accounts/private/v1/form/account/post``` returns the response:  
```json
{
    "data": [
        {
            "name": "number", // field name
            "dataType": "string", // field data type
            "validators": [  // list of validators
                // this validator will be applyed if typeId field has value 7 or 8
                {
                    "conditions": // validator can be depends from another field value
                        {
                            "fieldName": "typeId", // the validator depends from this field name
                            "values": [
                                7,
                                8
                            ],
                            "type": "in" // type of comparison
                        }
                    ],
                    "options": {
                        "value": 45 // validator options
                    },
                    "name": "max" // validator name
                }
            ]
        }
    ]
}
```

## Wallet Accounts Helm chart configuration

For usage examples and tips see [this article](https://velmie.atlassian.net/wiki/spaces/WAL/pages/52004603/Wallet-+Helm+charts+getting+started).

The following table lists the configurable parameters of the wallet-accounts chart, and their default values.

| Parameter                      | Description                                                                                                                      | Default                                 |
|--------------------------------|----------------------------------------------------------------------------------------------------------------------------------|:---------------------------------------:|
| service.type                   | The type of a service e.g. ClusterIp, NodePort, LoadBalancer                                                                     | ClusterIp                               |
| service.ports.public           | Application public API port.                                                                                                     | 10308                                   |
| service.ports.rpc              | Application RPC port.                                                                                                            | 12308                                   |
| service.ports.unsafeExposeRPC  | Forces to expose RPC port even if service.type other than ClusterIp                                                              | false                                   |
| service.selectors              | List of additional selectors                                                                                                     | {}                                      |
| containerPorts                 | List of ports that should be exposed on application container but in the service object.                                         | []                                      |
| containerLivenessProbe.enabled | Determines whether liveness probe should be performed on a pod.                                                                  |                                         |
| containerLivenessProbe.failureThreshold | Number of requests that should be failed in order to treat container unhealthy                                          | 5                                       |
| containerLivenessProbe.periodSeconds | Number of seconds between check requests.                                                                                  | 15                                      |
| appApiPathPrefix               | API prefix path. Used with internal health check functionality.                                                                  | accounts                              |
| mysqlAdmin.user                | Privileged database user name. Used in order to create DB schema and user. Required if hooks.dbInit.enabled=true.                |                                         |
| mysqlAdmin.password            | Privileged database user password.                                                                                               |                                         |
| hooks.dbInit.enabled           | Enabled database init job.                                                                                                       | false                                   |
| hooks.dbInit.createSchema      | Determines whether to create database schema. Depends on hooks.dbInit.enabled                                                    | true                                    |
| hooks.dbInit.createUser        | Determines whether to create database user that will be restricted to only use specified database schema.                        | true                                    |
| hooks.dbMigration.enabled      | Determines whether to run database migrations.                                                                                   |                                         |
| ingress.enabled                | Determines whether to create ingress resource for the service.                                                                   | true                                    |
| ingress.annotations            | List of additional annotations for the ingress.                                                                                  | {"kubernetes.io/ingress.class": "nginx"}|
| ingress.tls.enabled            | Determines whether TLS (https) connection should be set.                                                                         | false                                   |
| ingress.tls.host               | Host name that is covered by a certificate. This value is required if ingress.tls.enabled=true.                                  |                                         |
| ingress.tls.secretName         | [Kubernetes secret](https://kubernetes.io/docs/concepts/services-networking/ingress/#tls) name where TLS certificate is stored.  |                                         |
| appEnv.corsMethods             | Access-Control-Allow-Methods header that will be returned by the application.                                                    | GET,POST,PUT,OPTIONS                    |
| appEnv.corsOrigins             | Access-Control-Allow-Origin header that will be returned by the application.                                                     | *                                       |
| appEnv.corsHeaders             | Access-Control-Allow-Headers header that will be returned by the application.                                                    | *                                       |
| appEnv.dbHost                  | Database host to which application will be connected                                                                             | mysql                                   |
| appEnv.dbPort                  | Application database port.                                                                                                       | 3306                                    |
| appEnv.dbUser                  | Application database user.                                                                                                       |                                         |
| appEnv.dbName                  | Application database name.                                                                                                       |                                         |
| appEnv.dbDebugMode             | Whether database queries should be logged. Debugging mode.                                                                       | false                                   |
| image.repository               | What docker image to deploy.                                                                                                     | 360021420270.dkr.ecr.eu-central-1.amazonaws.com/velmie/wallet-currencies |
| image.pullPolicy               | What image pull policy to use.                                                                                                   | IfNotPresent                             |
| image.tag                      | What docker image tag to use.                                                                                                    | {Chart.yaml - appVersion}                |
| image.dbMigrationRepository    | What docker image to run in order to execute database migrations. By default the value if image.repository + "-db-migration"     | {image.tag}-db-migration                 |
| image.dbMigrationTag           | What docker image tag should be used for the db migration image.                                                                 | Same as image.tag                        |
| imagePullSecrets               | List of secrets which contain credentials to private docker repositories.                                                        | []                                       |
| nameOverride                   | Override this chart name.                                                                                                        | wallet-accounts                        |
| fullnameOverride               | Override this chart full name. By default it is composed from release name and the chart name.                                   | {releaseName}-{chartName}                |
| serviceAccount.create          | Whether Kubernetes service account resource should be created.                                                                   | false                                    |
| serviceAccount.annotations     | Annotations to add to the service account                                                                                        | {}                                       |
| serviceAccount.name            | The name of the service account to use. If not set and create is true, a name is generated using the fullname template.          | See description                          |
| podAnnotations                 | Kubernetes pod annotations.                                                                                                      | {}                                       |
| securityContext                | A security context defines privilege and access control settings for a Pod or Container. [See details](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/) | {} |
| resources                      | Limit Pod computing resources. [See details](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/)             | {}                                       |
| autoscaling.enabled            | Determines whether autoscaling functionality is enabled.                                                                         | false                                    |
| autoscaling.minReplicas        | [See details](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale-walkthrough/)                            | 1                                        |
| autoscaling.maxReplicas        | [See details](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale-walkthrough/)                            | 5                                        |
| autoscaling.targetCPUUtilizationPercentage | [See details](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale-walkthrough/)                | 80                                       |
| nodeSelector                   | [See details](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#nodeselector)                             | {}                                       |
| tolerations                    | [See details](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/)                                     | []                                       |
| affinity                       | [See details](https://kubernetes.io/docs/tasks/configure-pod-container/assign-pods-nodes-using-node-affinity/)                   | {}                                       |

## Run the project with Tilt

[Tilt](https://tilt.dev/) automates all the steps from a code change to a new process: watching files, building container images, and bringing your environment up-to-date.

[Install Tilt](https://docs.tilt.dev/install.html)

See [this article](https://velmie.atlassian.net/wiki/spaces/WAL/pages/56001240/Running+core+services+with+Tilt) which explains how to work with Tilt regarding this project.
