print("Wallet Accounts")

load("ext://restart_process", "docker_build_with_restart")

cfg = read_yaml(
    "tilt.yaml",
    default = read_yaml("tilt.yaml.sample"),
)

local_resource(
    "accounts-build-binary",
    "make fast_build",
    deps = ["./cmd", "./internal"],
)
local_resource(
    "accounts-generate-protpbuf",
    "make gen-protobuf",
    deps = ["./rpc/accounts/accounts.proto"],
)

docker_build(
    "velmie/wallet-accounts-db-migration",
    ".",
    dockerfile = "Dockerfile.migrations",
    only = "migrations",
)
k8s_resource(
    "wallet-accounts-db-migration",
    trigger_mode = TRIGGER_MODE_MANUAL,
    resource_deps = ["wallet-accounts-db-init"],
)

wallet_accounts_options = dict(
    entrypoint = "/app/service_accounts",
    dockerfile = "Dockerfile.prebuild",
    port_forwards = [],
    helm_set = [],
)

if cfg["debug"]:
    wallet_accounts_options["entrypoint"] = "$GOPATH/bin/dlv --continue --listen :%s --accept-multiclient --api-version=2 --headless=true exec /app/service_accounts" % cfg["debug_port"]
    wallet_accounts_options["dockerfile"] = "Dockerfile.debug"
    wallet_accounts_options["port_forwards"] = cfg["debug_port"]
    wallet_accounts_options["helm_set"] = ["containerLivenessProbe.enabled=false", "containerPorts[0].containerPort=%s" % cfg["debug_port"]]

docker_build_with_restart(
    "velmie/wallet-accounts",
    ".",
    dockerfile = wallet_accounts_options["dockerfile"],
    entrypoint = wallet_accounts_options["entrypoint"],
    only = [
        "./build",
        "zoneinfo.zip",
    ],
    live_update = [
        sync("./build", "/app/"),
    ],
)
k8s_resource(
    "wallet-accounts",
    resource_deps = ["wallet-accounts-db-migration"],
    port_forwards = wallet_accounts_options["port_forwards"],
)

yaml = helm(
    "./helm/wallet-accounts",
    # The release name, equivalent to helm --name
    name = "wallet-accounts",
    # The values file to substitute into the chart.
    values = ["./helm/values-dev.yaml"],
    set = wallet_accounts_options["helm_set"],
)

k8s_yaml(yaml)
