.PHONY: clean all

go_source ?= $(shell find ./ -type f -name "*.go")  \
			 go.mod go.sum

app: ${go_source}
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -tags "lazycat_cloud sqlite"  -ldflags  "-s -w -X main.Version='${DATE}.${VERSION}'"

publish: build
	echo "发布完成"

build: app
	echo "构建完成"




