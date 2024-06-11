SHELL := /bin/bash
HOME_DIR ?= $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
BACK_DIR = $(HOME_DIR)/back
WEBAPP_DIR = $(HOME_DIR)/webapp
GO_PATH ?= $(HOME_DIR)/_go
GO_BUILD ?= GOPATH=$(GO_PATH) CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w"
GO_RUN ?= GOPATH=$(GO_PATH) go run
DOCKER_BUILD ?= DOCKER_BUILDKIT=1 BUILDKIT_PROGRESS=plain docker build
DOCKER_IMAGE_NAME ?= qwasa.net/qqoin.backend
ENVFILE ?= $(BACK_DIR)/qqoin.env
O5_BIN := $(shell which o5)

include $(ENVFILE)


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
	cd $(WEBAPP_DIR)/ && ENVFILE=$(realpath $(ENVFILE)) bash build.sh


deploy_demo: DEPLOY_HOST ?= $(QQOIN_DEPLOY_HOST)
deploy_demo: TEMP_FILE := $(shell mktemp --dry-run)
deploy_demo: build

	ssh $(DEPLOY_HOST) ' \
	useradd -d /home/qqoin -m -g nogroup -s /bin/bash qqoin; \
	grep qqoin /etc/passwd; \
	systemctl stop qqoin-backend.service; \
	'

	scp $(BACK_DIR)/src/qqoin.backend $(DEPLOY_HOST):/home/qqoin/qqoin.backend
	scp -r webapp $(DEPLOY_HOST):/home/qqoin/

	scp $(ENVFILE) $(DEPLOY_HOST):/home/qqoin/qqoin.env
	scp misc/systemd-qqoin-backend.service $(DEPLOY_HOST):/etc/systemd/system/qqoin-backend.service
	$(O5_BIN) -dd $(ENVFILE) -start '%%' -end '%%' -i misc/nginx-qqoin.conf.template | tee '$(TEMP_FILE)'
	scp $(TEMP_FILE) $(DEPLOY_HOST):/etc/nginx/sites-enabled/qqoin.conf
	rm -f '$(TEMP_FILE)'

	ssh $(DEPLOY_HOST) '\
	whoami; \
	chmod -v 700 /home/qqoin/qqoin.backend /home/qqoin/qqoin.env; \
	mkdir -pv /home/qqoin/logs /home/qqoin/db; \
	chmod -Rv 755 /home/qqoin/logs /home/qqoin/db /home/qqoin/webapp; \
	chown -v qqoin:nogroup /home/qqoin/qqoin.backend /home/qqoin/qqoin.env /home/qqoin/logs /home/qqoin/webapp /home/qqoin/db; \
	systemctl daemon-reload; \
	systemctl enable qqoin-backend; systemctl start qqoin-backend.service; \
	systemctl restart nginx;\
	'

	curl -i https://qqoin.$(QQOIN_WEB_BASE_HOST)/ && echo "."
	curl -i https://qqoin-api.$(QQOIN_WEB_BASE_HOST)/api/ping/ && echo "."


deploy_tghook:
	curl -s -X POST \
	-F "url=https://qqoin-api.$(QQOIN_WEB_BASE_HOST)/api/tghook/" \
	-F "secret_token=$(QQOIN_BOT_SECRET_TOKEN)" \
	"https://api.telegram.org/bot$(QQOIN_BOT_TOKEN)/setWebhook"


show_tghook:
	curl -s -X POST "https://api.telegram.org/bot$(QQOIN_BOT_TOKEN)/getMyName" | json_pp
	curl -s -X POST "https://api.telegram.org/bot$(QQOIN_BOT_TOKEN)/getWebhookInfo" | json_pp
	curl -s -X POST "https://api.telegram.org/bot$(QQOIN_BOT_TOKEN)/getMyDescription" | json_pp
	curl -s -X POST "https://api.telegram.org/bot$(QQOIN_BOT_TOKEN)/getChatMenuButton" | json_pp
