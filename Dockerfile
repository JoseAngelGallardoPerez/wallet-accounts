FROM golang:1.14-alpine AS builder

RUN apk update && apk add --no-cache git mercurial openssh make

ARG REPOSITORY_PRIVATE_KEY

ARG GOOS=linux
ENV GO111MODULE=on
ENV GOPRIVATE=github.com/Confialink

WORKDIR $GOPATH/src/velmie/wallet-accounts

RUN mkdir -p ~/.ssh && umask 0077 && echo "${REPOSITORY_PRIVATE_KEY}" > ~/.ssh/id_rsa \
	&& git config --global url."git@github.com:Confialink".insteadOf https://github.com/Confialink \
	&& ssh-keyscan bitbucket.org >> ~/.ssh/known_hosts \
	&& ssh-keyscan github.com >> ~/.ssh/known_hosts

COPY . .

RUN make build

FROM alpine:3.11

RUN apk add wkhtmltopdf ca-certificates tzdata
RUN apk add --no-cache --virtual .build-deps msttcorefonts-installer \
    && update-ms-fonts \
    && fc-cache -f \
    && apk del .build-deps

WORKDIR /app

COPY --from=builder /go/src/velmie/wallet-accounts/build/service_accounts /app/service_accounts

COPY zoneinfo.zip /app/zoneinfo.zip
ENV ZONEINFO /app/zoneinfo.zip

ENTRYPOINT ["/app/service_accounts"]
