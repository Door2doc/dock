COMMIT =	$(shell git rev-list -1 HEAD | head -c 8)
NOW = 		$(shell date +"%Y-%m-%d %H:%M:%S")
CLEAN =		$(shell git status -s)

.PHONY:	release
release:
ifneq "$(CLEAN)" ""
	$(error There are uncommitted changes: "$(CLEAN)")
endif
ifeq "$(VERSION)" ""
	$(error There is no version specified)
endif
	git push origin master
	$(MAKE) compile
	hub release create -d -a d2d-upload_windows_amd64.exe -m"$(VERSION)" $(VERSION)

.PHONY:	compile
compile:	generate
	docker run --rm \
		-v "$(shell pwd)":/gopath/src/github.com/publysher/d2d-uploader \
		-w /gopath/src/github.com/publysher/d2d-uploader tcnksm/gox:1.10.3 \
		gox \
			-osarch="windows/amd64 darwin/amd64" \
			-ldflags '-X "main.GitCommit=$(COMMIT)" \
					  -X main.Version=$(VERSION) \
					  -X "main.Built=$(NOW)" \
					  ' \
			./...

.PHONY:	clean
clean:
	rm -f d2d-upload_darwin_amd64 d2d-upload_windows_amd64.exe

.PHONY:	test
test:
	go test ./...

.PHONY:	generate
generate:
	go get -u github.com/mjibson/esc

	$(shell go env GOPATH)/bin/esc -o pkg/uploader/web/assets.go -pkg web -prefix pkg/uploader/web/resources pkg/uploader/web/resources