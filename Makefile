USERSPACE=Confialink
NAMESPACE=wallet
SERVICE=accounts

APP=./build/service_${SERVICE}
PROJECT?=github.com/${USERSPACE}/${NAMESPACE}-${SERVICE}
DATE := $(shell date +'%Y.%m.%d %H:%M:%S')

ndef = $(if $(value $(1)),,$(error required environment variable $(1) is not set))

ifndef COMMIT
	COMMIT := $(shell git rev-parse HEAD)
endif

ifndef TAG
	TAG = $(shell git describe --exact-match --tags $(git log -n1 --pretty='%h') 2>/dev/null)
endif

GOOS?=linux
DOCKER_TAG?=wallet-accounts
GO111MODULE?=on
GOPRIVATE?=github.com/Confialink

show:
	@echo ${PROJECT}
	@echo ${DATE}

fast_build:
	CGO_ENABLED=0 GOOS=${GOOS} go build \
		-gcflags "all=-N -l" \
		-ldflags '-X "${PROJECT}/version.DATE=${DATE}" -X ${PROJECT}/version.COMMIT=${COMMIT} -X ${PROJECT}/version.TAG=${TAG}' \
		-o ${APP} ./cmd/main.go

build: clean
	CGO_ENABLED=0 GOOS=${GOOS} go build -a -installsuffix cgo \
		-ldflags '-s -w -X "${PROJECT}/version.DATE=${DATE}" -X ${PROJECT}/version.COMMIT=${COMMIT} -X ${PROJECT}/version.TAG=${TAG}' \
		-o ${APP} ./cmd/main.go

docker-build:
	$(call ndef,REPOSITORY_PRIVATE_KEY)
	docker build . --build-arg REPOSITORY_PRIVATE_KEY --build-arg TAG=${TAG} -t ${DOCKER_TAG}

gen-protobuf:
	protoc --proto_path=. --go_out=. --twirp_out=. rpc/accounts/accounts.proto
	protoc --proto_path=. --go_out=. --twirp_out=. rpc/limit/limit.proto

clean:
	@[ -f ${APP} ] && rm -f ${APP} || true
