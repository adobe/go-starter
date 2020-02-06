.PHONY: clean build install

COMMIT?=${BUILDCOMMIT}
VERSION?=${BUILDTAG}

# enable cgo because it's required by OSX keychain library
CGO_ENABLED=1

# enable go modules
GO111MODULE=on

export CGO_ENABLED GO111MODULE

dep:
	go get ./...

test:
	go test ./...

lint:
	golangci-lint run

go-starter: cmd/go-starter/* pkg/*
	go build -ldflags "-X main.version=${VERSION} -X main.commit=${COMMIT} ${BUILDLDFLAGS}" ${BUILDARGS} \
		-o ${BUILDOUTPREFIX}go-starter cmd/go-starter/main.go

go-starter-replace: cmd/go-starter-replace/* pkg/*
	go build -ldflags "-X main.version=${VERSION} -X main.commit=${COMMIT} ${BUILDLDFLAGS}" ${BUILDARGS} \
		-o ${BUILDOUTPREFIX}go-starter-replace cmd/go-starter-replace/main.go

go-starter-github: cmd/go-starter-github/* pkg/*
	go build -ldflags "-X main.version=${VERSION} -X main.commit=${COMMIT} ${BUILDLDFLAGS}" ${BUILDARGS} \
		-o ${BUILDOUTPREFIX}go-starter-github cmd/go-starter-github/main.go

go-starter-drone: cmd/go-starter-drone/* pkg/*
	go build -ldflags "-X main.version=${VERSION} -X main.commit=${COMMIT} ${BUILDLDFLAGS}" ${BUILDARGS} \
		-o ${BUILDOUTPREFIX}go-starter-drone cmd/go-starter-drone/main.go

clean:
	rm ${BUILDOUTPREFIX}go-starter* 2> /dev/null || exit 0

build: go-starter go-starter-replace go-starter-github go-starter-drone

install: build
	cp ${BUILDOUTPREFIX}go-starter* /usr/local/bin
