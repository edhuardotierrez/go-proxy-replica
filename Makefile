.PHONY : all

# other
cat := $(if $(filter $(OS),Windows_NT),type,cat)
MY_VERSION := $(shell $(cat) VERSION)

all: go

go:
	go build -ldflags="-X 'main.Version=v$(MY_VERSION)'" -o go-proxy-replica$(SUFFIX)

go_linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-X 'main.Version=v$(MY_VERSION)'" -o go-proxy-replica$(SUFFIX)

vendor:
	go env -w GO111MODULE=on
	go mod tidy -go=1.16 && go mod tidy -go=1.17
	go mod vendor
