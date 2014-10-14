VERSION    := $(shell git tag -l v*.*.* | sort -r -n -t . | head -1)
VERSION    := v0.0.0
GITSHA1    := $(shell git rev-parse --short HEAD)
GITROOT    := $(shell git rev-parse --show-toplevel)

GOFILES    := $(shell find . -type f -and -name "*.go")
GOFLAGS    := -ldflags "-X main.Version $(VERSION) -X main.GitSHA $(GITSHA1)"
GOHOSTOS   := $(shell go env GOHOSTOS)
GOHOSTARCH := $(shell go env GOHOSTARCH)

NAME             := $(notdir $(PWD))
DOCKER_SRC       := /go/src/$(subst $(GOPATH)/src/,,$(PWD))
DOCKER_IMAGE     := $(NAME)-golang
DOCKER_CONTAINER := $(NAME)-build

# Support boot2docker
ifeq (0, $(shell which boot2docker >/dev/null; echo $$?))
	export DOCKER_HOST := $(shell boot2docker shellinit | sed "s/[^=]*=//")
endif

# Template Dockerfile
define DOCKER_FILE
# Dockerfile to cross compile `$(NAME)`

FROM golang:1.3-cross

ADD . /go/src/$(subst $(GOPATH)/src/,,$(GITROOT))
WORKDIR $(DOCKER_SRC)

# Download (but not install) dependencies
RUN go get -d -v ./...

# Turn this container into `make`
ENTRYPOINT ["/usr/bin/make"]
endef

default: $(shell echo bin/$(VERSION)/{darwin,linux}-{amd64,386})
default: $(shell echo bin/$(VERSION)/windows-{amd64,386}.exe)
	@true

# Automatically generate Dockerfile for this golang application
Dockerfile: export DOCKER_FILE := $(DOCKER_FILE)
Dockerfile:
	@echo "$$DOCKER_FILE"> Dockerfile

# Build Docker container
docker.image: Dockerfile
	# Docker contexts are broken, circumvent this
	cp Dockerfile $(GITROOT)
	docker build -t "$(DOCKER_IMAGE)" $(GITROOT)
	rm $(GITROOT)/Dockerfile

	# Touch docker.image to make make play along
	touch $@

# Compile the program using the above container
bin/%: docker.image
	mkdir -p $(@D)
	docker run --name "$(DOCKER_CONTAINER)-$(*F)" "$(DOCKER_IMAGE)" build/$*
	docker cp "$(DOCKER_CONTAINER)-$(*F)":"$(DOCKER_SRC)/build/$*" $(@D) || true
	docker rm "$(DOCKER_CONTAINER)-$(*F)"

# Native Go build per OS-ARCH combo, inside container.
build/%: $(GOFILES)
	mkdir -p $(@D)
	GOOS=$(subst -, GOARCH=,$(basename $(*F))) go build $(GOFLAGS) -o $@

# Remove built binaries and Docker containers
clean:
	$(eval NAMES := $(shell docker ps -a | grep -o "$(NAME)-build-.*"))
	$(if $(NAMES),docker rm $(NAMES))
	go clean -i
	rm -fr bin build docker.image $(GITROOT)/Dockerfile

# Build for this platform using docker
docker-install: bin/$(VERSION)/$(GOHOSTOS)-$(GOHOSTARCH)
	@true

# Run the resulting binary
docker-run: docker-install
	@bin/$(VERSION)/$(GOHOSTOS)-$(GOHOSTARCH)

# Bypass docker and build for this platform directly
install: $(GOFILES)
	@go install $(GOFLAGS)

# Run the resulting binary
run: install
	@$(NAME)

# Virtual targets
.PHONY: default clean docker-install docker-run install run
