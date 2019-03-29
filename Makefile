COMMIT =	$(shell git rev-list -1 HEAD | head -c 8)
NOW = 		$(shell date +"%Y-%m-%d %H:%M:%S")
CLEAN =		$(shell git status -s)
SOURCES = 	$(shell find . -name '*.go')

.PHONY:	release
release:
ifneq "$(CLEAN)" ""
	$(error There are uncommitted changes: "$(CLEAN)")
endif
ifeq "$(VERSION)" ""
	$(error There is no version specified)
endif
	git push origin master
	$(MAKE) Door2doc_Upload_Service_$(VERSION).exe
	$(MAKE) -C doc handleiding.pdf
	hub release create -d -a "Door2doc_Upload_Service_$(VERSION).exe#Windows installer" -m"$(VERSION)" $(VERSION)
	hub release edit -d -a "doc/handleiding.pdf#Handleiding" -m"$(VERSION)" $(VERSION)

installer: Door2doc_Upload_Service_$(VERSION).exe 

Door2doc_Upload_Service_$(VERSION).exe: d2d-upload_windows_amd64.exe installer.nsi
	docker run -it -v "$(shell pwd)":/app hp41/nsis:3.01-1 -DVERSION=$(VERSION) /app/installer.nsi
	mv _installer.exe $@


d2d-upload_windows_amd64.exe:	$(SOURCES)
	$(MAKE) generate
	docker run --rm \
		-v "$(shell pwd)":/gopath/src/github.com/publysher/d2d-uploader \
		-w /gopath/src/github.com/publysher/d2d-uploader tcnksm/gox:1.10.3 \
		gox \
			-osarch="windows/amd64" \
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

	$(shell go env GOPATH)/bin/esc -o pkg/uploader/assets/assets.go -pkg assets -prefix pkg/uploader/assets/resources pkg/uploader/assets/resources

