HOME_DIR ?= $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
BACK_DIR = $(HOME_DIR)/back
GO_PATH ?= $(HOME_DIR)/_go
GO_BUILD ?= GOPATH=$(GO_PATH) CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w"
GO_RUN ?= GOPATH=$(GO_PATH) go run
DOCKER_BUILD ?= DOCKER_BUILDKIT=1 BUILDKIT_PROGRESS=plain docker build
DOCKER_IMAGE_NAME ?= qwasa.net/qqoin.backend
ENVFILE ?= $(BACK_DIR)/qqoin.env


hello:
	@true


build: back.build webapp.build


run: back.run


back.build:
	cd $(BACK_DIR)/src && \
	$(GO_BUILD) .


back.run:
	cd $(BACK_DIR)/src && \
	$(GO_RUN) .


back.docker-image-build:
	cd $(BACK_DIR) && \
	$(DOCKER_BUILD) . -t $(DOCKER_IMAGE_NAME)


back.docker-container-run:
	docker run --publish-all --interactive --rm --env-file=$(ENVFILE) $(DOCKER_IMAGE_NAME)


webapp.build:
	true


demo_deploy: DEPLOY_HOST ?= root.qqoin.cc
demo_deploy: build

	@ssh $(DEPLOY_HOST) ' \
	useradd -d /home/qqoin -m -g nogroup -s /bin/bash qqoin; \
	grep qqoin /etc/passwd; \
	systemctl stop qqoin-backend.service; \
	'

	@scp $(BACK_DIR)/src/qqoin.backend $(DEPLOY_HOST):/home/qqoin/qqoin.backend
	@scp -r webapp $(DEPLOY_HOST):/home/qqoin/webapp

	@scp $(ENVFILE) $(DEPLOY_HOST):/home/qqoin/qqoin.env
	@scp misc/systemd-qqoin-backend.service $(DEPLOY_HOST):/etc/systemd/system/qqoin-backend.service
	@scp misc/nginx-qqoin.conf $(DEPLOY_HOST):/etc/nginx/sites-enabled/qqoin.conf

	@ssh $(DEPLOY_HOST) '\
	whoami; \
	chmod -v 700 /home/qqoin/qqoin.backend /home/qqoin/qqoin.env; \
	mkdir -pv /home/qqoin/logs /home/qqoin/db; \
	chmod -Rv 755 /home/qqoin/logs /home/qqoin/db /home/qqoin/webapp; \
	chown -v qqoin:nogroup /home/qqoin/qqoin.backend /home/qqoin/qqoin.env /home/qqoin/logs /home/qqoin/webapp /home/qqoin/db; \
	systemctl daemon-reload; systemctl enable qqoin-backend; systemctl start qqoin-backend.service; \
	systemctl restart nginx;\
	'