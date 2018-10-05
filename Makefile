DOTENV := ./.env
DOTENV_EXISTS := $(shell [ -f $(DOTENV) ] && echo 0 || echo 1 )

ifeq ($(DOTENV_EXISTS), 0)
	include $(DOTENV)
	export $(shell sed 's/=.*//' .env)
endif

GO15VENDOREXPERIMENT = 1
OSXCROSS_NO_INCLUDE_PATH_WARNINGS = 1
VERSION = v0.0.1

NAME	 := qicoo-api
TARGET	 := bin/$(NAME)
DIST_DIRS := find * -type d -exec
SRCS	:= $(shell find . -type f -name '*.go')
LDFLAGS := -ldflags="-s -X \"main.version=$(VERSION)\""

$(TARGET): $(SRCS)
	golint src/${NAME}.go
	go build $(OPTS) $(LDFLAGS) -o bin/$(NAME) src/${NAME}.go

.PHONY: create-dotenv
create-dotenv:
	@if [ ! -f $(DOTENV) ]; \
		then\
		echo 'Create .env file.' ;\
		echo 'DB_USER=root' >> ./.env ;\
		echo 'DB_PASSWORD=root' >> ./.env ;\
		echo 'DB_URL=localhost:3306' >> ./.env ;\
		echo 'REDIS_URL=localhost:6379' >> ./.env ;\
		echo 'IS_TRAVISENV=' >> ./.env ;\
	else \
		echo Not Work. ;\
	fi

.PHONY: test-main
test-main:
	go test -v ./src/qicoo-api_test.go

.PHONY: test-sql
test-sql:
	go test -v ./src/sql/mysql_test.go ./src/sql/mysql.go

.PHONY: test-list
test-list:
	@if test "$(IS_TRAVISENV)" = "true" ;\
		then \
		go test -v ./src/handler/list_test.go ./src/handler/list.go ./src/handler/sync.go -run TestGetQuestionListInTheTravis ;\
	else \
		go test -v ./src/handler/list_test.go ./src/handler/list.go ./src/handler/sync.go -run TestGetQuestionListInTheLocal ;\
	fi

.PHONY: test
test: clean-test test-sql test-list test-main

.PHONY: install
install:
	go install $(LDFLAGS)

.PHONY: clean-test
clean-test:
	go clean --testcache

.PHONY: clean
clean:
	rm -rf bin/*

.PHONY: clean-all
clean-all: clean
	rm -rf dist/*

.PHONY: run
run:
	go run src/$(NAME).go

.PHONY: upde
upde:
	dep ensure --update

.PHONY: deps
dep:
	dep ensure

.PHONY: dep-install
dep-install:
	go get -u github.com/golang/dep/cmd/dep

.PHONY: cross-build
cross-build: deps
	for os in darwin linux windows; do \
		for arch in amd64 386; do \
			GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build -a -tags netgo -installsuffix netgo $(LDFLAGS) -o dist/$(NAME)-$$os-$$arch/$(NAME); \
		done; \
	done

.PHONY: dist
dist:
	cd dist && \
		$(DIST_DIRS) cp ../LICENSE {} \; && \
		$(DIST_DIRS) cp ../README.md {} \; && \
		$(DIST_DIRS) tar -zcf {}-$(VERSION).tar.gz {} \; && \
		$(DIST_DIRS) zip -r {}-$(VERSION).zip {} \; && \
		cd ..

