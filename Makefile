LDFLAGS = -ldflags "-s -w"
TAGS = -tags nosystemd

GOOS=linux
GOARCH=amd64

.PHONY: all
all: build

.PHONY: clean
clean:
	rm -f postfix_exporter cover.out

.PHONY: test
test:
	go test -coverprofile cover.out -count=1 -race -p 4 -v ./...

.PHONY: build
build:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(LDFLAGS) $(TAGS) .

.PHONY: docker
docker: build
	docker build --platform linux/amd64 -t numo68/postfix_exporter:latest .
