COMMIT =	$(shell git rev-list -1 HEAD | head -c 8)
NOW = 		$(shell date +"%Y-%m-%d %H:%M:%S")
CLEAN =		$(shell git status -s)
SOURCES = 	$(shell find . -name '*.go')
ESC = 		$(shell go env GOPATH)/bin/esc

.PHONY:	release
release:
ifneq "$(CLEAN)" ""
	$(error There are uncommitted changes: "$(CLEAN)")
endif
ifeq "$(VERSION)" ""
	$(error There is no version specified)
endif
	git push origin master
	$(MAKE) clean
	$(MAKE) Dock_$(VERSION).exe
	$(MAKE) -C doc handleiding.pdf
	gh release create $(VERSION) --prerelease --title "$(VERSION)" --notes ""
	gh release upload $(VERSION) "Dock_$(VERSION).exe#Windows installer" "doc/handleiding.pdf#Handleiding"

installer: Dock_$(VERSION).exe

Dock_$(VERSION).exe: d2d-upload_windows_amd64.exe installer.nsi
	sync
	-docker run -it -w /app -v "$(shell pwd)":/app hp41/nsis:3.01-1 -DVERSION=$(VERSION) installer.nsi
	docker run -it -w /app -v "$(shell pwd)":/app hp41/nsis:3.01-1 -DVERSION=$(VERSION) installer.nsi
	mv _installer.exe $@


d2d-upload_windows_amd64.exe:	$(SOURCES)
	$(MAKE) generate
	GOOS=windows GOARCH=amd64 go build -o $@ -ldflags '-X "main.GitCommit=$(COMMIT)" -X "main.Version=$(VERSION)" -X "main.Built=$(NOW)"' ./cmd/d2d-upload

.PHONY:	clean
clean:
	rm -f *.exe

.PHONY:	test
test:
	go test ./...


$(ESC):
	go get -u github.com/mjibson/esc
	git restore go.mod go.sum



.PHONY:	generate
generate:	$(ESC)
	$(ESC) -o pkg/uploader/assets/assets.go -pkg assets -prefix pkg/uploader/assets/resources pkg/uploader/assets/resources
