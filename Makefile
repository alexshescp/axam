VERSION ?= 1.1.0
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: test vet race build clean deb

test:
	go test ./...

vet:
	go vet ./...

race:
	go test -race ./...

build:
	CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o axam ./cmd/axam

dist:
	mkdir -p dist
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o dist/axam_linux_amd64 ./cmd/axam
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags "$(LDFLAGS)" -o dist/axam_linux_arm64 ./cmd/axam
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o dist/axam_darwin_amd64 ./cmd/axam
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags "$(LDFLAGS)" -o dist/axam_darwin_arm64 ./cmd/axam

deb: dist
	go run packaging/mkdeb.go -bin dist/axam_linux_amd64 -arch amd64 -version $(VERSION) -out dist/axam_$(VERSION)_amd64.deb
	go run packaging/mkdeb.go -bin dist/axam_linux_arm64 -arch arm64 -version $(VERSION) -out dist/axam_$(VERSION)_arm64.deb

clean:
	rm -rf dist axam axam.exe
