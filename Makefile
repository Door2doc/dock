COMMIT =	$(shell git rev-list -1 HEAD | head -c 8)
NOW = 		$(shell date +"%Y-%m-%d %H:%M:%S")
GOVERSION =	$(shell go version)

.PHONY:	compile
compile:
	docker run --rm \
		-v "$(shell pwd)":/gopath/src/github.com/publysher/d2d-uploader \
		-w /gopath/src/github.com/publysher/d2d-uploader tcnksm/gox:1.10.3 \
		gox \
			-osarch="windows/amd64 darwin/amd64" \
			-ldflags '-X "main.GitCommit=$(COMMIT)" \
					  -X main.Version=todo \
					  -X "main.Built=$(NOW)" \
					  ' \
			./...

.PHONY:	clean
clean:
	rm -f d2d-upload_darwin_amd64 d2d-upload_windows_amd64.exe

.PHONY:	test
test:
	go test ./...