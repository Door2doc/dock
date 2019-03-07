.PHONY:	compile
compile:
	docker run --rm -v "$(shell pwd)":/usr/src/d2d -w /usr/src/d2d tcnksm/gox:1.10.3 gox -osarch="windows/amd64 darwin/amd64" ./...

.PHONY:	clean
clean:
	go clean -i
	rm -f d2d-upload_darwin_amd64 d2d-upload_windows_amd64.exe

.PHONY:	test
test:
	go test ./...