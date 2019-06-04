PHONY: deps clean build

ENVIRONMENT        ?= prod
PROJECT            ?=  twitter-api
STACK_NAME         ?=  stack-twitter-api
ARTIFACTS_BUCKET   ?=  ba78-twitter-lambda
AWS_DEFAULT_REGION ?= eu-north-1

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

clean: 
	rm -rf ./twitter-bot1/dist
	rm deploy.yaml

build:
	mkdir -p twitter-bot1/dist
	cd twitter-bot1; GOOS=linux GOARCH=amd64 go build -o dist/twitter-bot1 ./
	cd twitter-bot1/dist; zip twitter-bot1.zip twitter-bot1

deploy:
	@make build
	@make test
	$(call sam_package)
	$(call sam_deploy)
	make clean

test:
	cd twitter-bot1; go test -v

delete-stack:
	aws cloudformation delete-stack --stack-name $(STACK_NAME)