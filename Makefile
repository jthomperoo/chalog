VERSION=development

default:
	@echo "=============Building binaries============="

	# Linux 386
	CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -ldflags="-X 'main.Version=$(VERSION)'" -o dist/linux_386/chalog main.go
	cp LICENSE dist/linux_386/LICENSE

	# Linux amd64
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-X 'main.Version=$(VERSION)'" -o dist/linux_amd64/chalog main.go
	cp LICENSE dist/linux_amd64/LICENSE

	# Linux arm
	CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -ldflags="-X 'main.Version=$(VERSION)'" -o dist/linux_arm/chalog main.go
	cp LICENSE dist/linux_arm/LICENSE

	# Linux arm64
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-X 'main.Version=$(VERSION)'" -o dist/linux_arm64/chalog main.go
	cp LICENSE dist/linux_arm64/LICENSE

	# Darwin amd64
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-X 'main.Version=$(VERSION)'" -o dist/darwin_amd64/chalog main.go
	cp LICENSE dist/darwin_amd64/LICENSE

	# Darwin arm64
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-X 'main.Version=$(VERSION)'" -o dist/darwin_arm64/chalog main.go
	cp LICENSE dist/darwin_arm64/LICENSE

	# Windows 386
	CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -ldflags="-X 'main.Version=$(VERSION)'" -o dist/windows_386/chalog.exe main.go
	cp LICENSE dist/windows_386/LICENSE

	# Windows amd64
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-X 'main.Version=$(VERSION)'" -o dist/windows_amd64/chalog.exe main.go
	cp LICENSE dist/windows_amd64/LICENSE

zip:
	@echo "=============Zipping binaries============="
	zip -r -j dist/chalog_linux_386.zip dist/linux_386
	zip -r -j dist/chalog_linux_amd64.zip dist/linux_amd64
	zip -r -j dist/chalog_linux_arm.zip dist/linux_arm
	zip -r -j dist/chalog_linux_arm64.zip dist/linux_arm64
	zip -r -j dist/chalog_darwin_amd64.zip dist/darwin_amd64
	zip -r -j dist/chalog_darwin_arm64.zip dist/darwin_arm64
	zip -r -j dist/chalog_windows_386.zip dist/windows_386
	zip -r -j dist/chalog_windows_amd64.zip dist/windows_amd64

test: unit_test integration_test

lint:
	@echo "=============Linting============="
	golint -set_exit_status ./...

beautify:
	@echo "=============Beautifying============="
	gofmt -s -w .
	go mod tidy

integration_test:
	@echo "=============Running integration tests============="
	go clean -testcache && go test ./... -tags=integration

unit_test:
	@echo "=============Running unit tests============="
	go test ./... -tags=unit -cover
