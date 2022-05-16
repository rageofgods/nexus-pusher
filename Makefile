.PHONY:
.SILENT:

#VERSION=`git describe --tags`
VERSION=`git rev-parse --short HEAD`
BUILD=`date +%FT%T%z`

build:
	go build -ldflags "-X main.Version=${VERSION} -X main.Build=${BUILD}" -o ./.bin/nexus cmd/nexus/main.go
docker_build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
		-ldflags "-X main.Version=${VERSION} -X main.Build=${BUILD} -w -s -extldflags '-static'" -a \
		-o nexus-pusher cmd/nexus/main.go
run: build
	./.bin/nexus
test:
	go test ./... -v
lint:
	golangci-lint run --enable gosec,bodyclose,dupl,unparam,wastedassign,unconvert,tagliatelle,rowserrcheck,predeclared,prealloc,nosprintfhostport,nonamedreturns,nilnil,nilerr,nakedret,misspell,makezero,lll,ireturn,ifshort,goprintffuncname,godox,gocyclo,gocritic,goconst,gochecknoinits,forcetypeassert,forbidigo,exportloopref,exhaustive,errorlint,errchkjson,durationcheck,dogsled,decorder,bidichk,asciicheck