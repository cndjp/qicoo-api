DOTENV := ./.env
DOTENV_EXISTS := $(shell [ -f $(DOTENV) ] && echo 0 || echo 1 )

ifeq ($(DOTENV_EXISTS), 0)
	include $(DOTENV)
	export $(shell sed 's/=.*//' .env)
endif

GO15VENDOREXPERIMENT = 1
OSXCROSS_NO_INCLUDE_PATH_WARNINGS = 1
VERSION = v0.0.1-dev

NAME	 := qicoo-api
TARGET	 := bin/$(NAME)
DIST_DIRS := find * -type d -exec
SRCS	:= $(shell find . -type f -name '*.go')
LDFLAGS := -ldflags="-s -X \"main.version=$(VERSION)\""

$(TARGET): $(SRCS)
	CGO_ENABLED=0 go build $(OPTS) $(LDFLAGS) -o bin/$(NAME) src/${NAME}.go

.PHONY: create-dotenv
create-dotenv:
	@if [ ! -f $(DOTENV) ]; \
		then\
		echo 'Create .env file.' ;\
		echo 'DB_USER=root' >> ./.env ;\
		echo 'DB_PASSWORD=root' >> ./.env ;\
		echo 'DB_URL=localhost:3306' >> ./.env ;\
		echo 'REDIS_URL=localhost:6379' >> ./.env ;\
		echo 'DOCKER_USERNAME=' >> ./.env ;\
		echo 'DOCKER_PASSWORD=' >> ./.env ;\
		echo 'TRAVIS=' >> ./.env ;\
	else \
		echo Not Work. ;\
	fi

.PHONY: test-main
test-main:
	go test -v ./src/qicoo-api_test.go

.PHONY: test-question-list
test-question-list:
	@if test "$(TRAVIS)" = "true" ;\
		then \
		go test -v ./src/handler/question-list_test.go \
		  ./src/handler/question-handler_test.go \
		  -run TestGetQuestionList;\
	else \
		docker run --name qicoo-api-test-mysql --rm -d -e MYSQL_ROOT_PASSWORD=my-secret-pw -p 3306:3306 mysql:5.6.27 --character-set-server=utf8mb4 --collation-server=utf8mb4_unicode_ci;\
		docker run --name qicoo-api-test-redis --rm -d -p 6379:6379 redis:4.0.10;\
		sleep 15;\
		go test -v ./src/handler/question-list_test.go \
		  ./src/handler/question-handler_test.go \
		  -run TestGetQuestionList;\
		docker kill qicoo-api-test-mysql;\
		docker kill qicoo-api-test-redis;\
	fi

.PHONY: test-question-create
test-question-create:
	@if test "$(TRAVIS)" = "true" ;\
		then \
		go test -v ./src/handler/question-create_test.go \
		  ./src/handler/question-handler_test.go \
		  -run TestCreateQuestion;\
	else \
		docker run --name qicoo-api-test-mysql --rm -d -e MYSQL_ROOT_PASSWORD=my-secret-pw -p 3306:3306 mysql:5.6.27 --character-set-server=utf8mb4 --collation-server=utf8mb4_unicode_ci;\
		docker run --name qicoo-api-test-redis --rm -d -p 6379:6379 redis:4.0.10;\
		sleep 15;\
		go test -v ./src/handler/question-create_test.go \
		  ./src/handler/question-handler_test.go \
		  -run TestCreateQuestion;\
		docker kill qicoo-api-test-mysql;\
		docker kill qicoo-api-test-redis;\
	fi

.PHONY: test-question-delete
test-question-delete:
	@if test "$(TRAVIS)" = "true" ;\
		then \
		go test -v ./src/handler/question-delete_test.go \
		  ./src/handler/question-create_test.go \
		  ./src/handler/question-handler_test.go \
		  -run TestDeleteQuestion;\
	else \
		docker run --name qicoo-api-test-mysql --rm -d -e MYSQL_ROOT_PASSWORD=my-secret-pw -p 3306:3306 mysql:5.6.27 --character-set-server=utf8mb4 --collation-server=utf8mb4_unicode_ci;\
		docker run --name qicoo-api-test-redis --rm -d -p 6379:6379 redis:4.0.10;\
		sleep 15;\
		go test -v ./src/handler/question-delete_test.go \
		  ./src/handler/question-create_test.go \
		  ./src/handler/question-handler_test.go \
		  -run TestDeleteQuestion;\
		docker kill qicoo-api-test-mysql;\
		docker kill qicoo-api-test-redis;\
	fi


test-question-like:
	@if test "$(TRAVIS)" = "true" ;\
		then \
		go test -v ./src/handler/question-like_test.go \
		  ./src/handler/question-handler_test.go \
		  -run TestQuestionLike;\
	else \
		docker run --name qicoo-api-test-mysql --rm -d -e MYSQL_ROOT_PASSWORD=my-secret-pw -p 3306:3306 mysql:5.6.27 --character-set-server=utf8mb4 --collation-server=utf8mb4_unicode_ci;\
		docker run --name qicoo-api-test-redis --rm -d -p 6379:6379 redis:4.0.10;\
		sleep 15;\
		go test -v ./src/handler/question-like_test.go \
		  ./src/handler/question-handler_test.go \
		  -run TestQuestionLike;\
		docker kill qicoo-api-test-mysql;\
		docker kill qicoo-api-test-redis;\
	fi


.PHONY: test
test: clean-test test-question-list test-question-create test-question-like test-main

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

.PHONY: lint
lint:
	golint src/${NAME}.go

.PHONY: golint-install
golint-install:
	go get -u golang.org/x/lint/golint

.PHONY: docker-build
docker-build:
	docker build -t cndjp/$(NAME):$(VERSION) .;\

.PHONY: docker-push
docker-push:
	echo "$(DOCKER_PASSWORD)" | docker login -u "$(DOCKER_USERNAME)" --password-stdin
	docker push cndjp/$(NAME):$(VERSION)

.PHONY: github-clone
github-clone:
	chmod +x ./github.sh
	./github.sh

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

