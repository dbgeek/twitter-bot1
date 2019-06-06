PHONY: deps clean build

ENVIRONMENT        ?= prod
PROJECT            ?= twitter-api
STACK_NAME         ?= stack-twitter-api
ARTIFACTS_BUCKET   ?= ba78-twitter-lambda
AWS_DEFAULT_REGION ?= eu-north-1

dirs = $(shell find * -type d -maxdepth 0)
baseDir = $(shell pwd)

sam_package = aws cloudformation package \
                --template-file sam.yaml \
                --output-template-file deploy.yaml \
                --s3-bucket $(ARTIFACTS_BUCKET)

sam_deploy = aws cloudformation deploy \
                --template-file deploy.yaml \
                --stack-name $(STACK_NAME) \
				--region $(AWS_DEFAULT_REGION) \
                --capabilities CAPABILITY_IAM \
                --no-fail-on-empty-changeset

deps:
	go get -u ./...

clean-all:
	@$(foreach dir,$(dirs), \
		echo $(dir); \
		cd $(baseDir); \
		rm -rf ./$(dir)/dist \
	)
	cd $(baseDir)

build-all:
	@$(foreach dir,$(dirs), \
		echo $(dir); \
		mkdir -p $(dir)/dist; \
		cd $(dir); GOOS=linux GOARCH=amd64 go build -o dist/$(dir) ./ ; \
		cd dist; zip $(dir).zip $(dir)  \
	)
	cd $(baseDir)

deploy:
	$(call sam_package)
	$(call sam_deploy)
	make clean

test-all:
	@$(foreach dir,$(dirs), \
		echo $(dir); \
		cd $(baseDir); \
		cd $(dir); \
		go test -v ; \
	)
	cd $(baseDir)

delete-stack:
	aws cloudformation delete-stack --stack-name $(STACK_NAME)