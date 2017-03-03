
TEST?=$$(go list ./... | grep -v '/vendor/')
GOFMT_FILES?=$$(find . -type f -name '*.go' | grep -v vendor)

build:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o rabbit-amazon-forwarder .

release: build
	docker build -t airhelp/rabbit-amazon-forwarder -f Dockerfile .

push: test release
	docker push airhelp/rabbit-amazon-forwarder

fmt:
	@echo 'run Go autoformat'
	@gofmt -w $(GOFMT_FILES)

# vet runs the Go source code static analysis tool `vet` to find
# any common errors.
vet:
	@echo 'run the code static analysis tool'
	@go tool vet -all $$(ls -d */ | grep -v vendor)

test: fmt vet
	@echo 'run the unit tests'
	@go test -cover $(TEST)

dev:
	go build
