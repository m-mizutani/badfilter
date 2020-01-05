STACK_CONFIG ?= stack.jsonnet
SAM_CONFIG ?= sam.jsonnet

CODE_DIR := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
CWD := ${CURDIR}
BINPATH := $(CWD)/build/updater $(CWD)/build/matcher
INTERNAL := $(CODE_DIR)/internal/*

TEMPLATE_FILE := template.json
SAM_FILE := sam.yml
BASE_FILE := $(CODE_DIR)/template.libsonnet
OUTPUT_FILE := $(CWD)/output.json

STACK_NAME := $(shell jsonnet $(STACK_CONFIG) | jq .StackName)
BUILD_OPT :=

ifdef TAGS
TAGOPT=--tags $(TAGS)
else
TAGOPT=
endif

all: $(OUTPUT_FILE)

clean:
	go clean
	rm -f $(BINPATH)

build: $(BINPATH)

$(CWD)/build/updater: $(CODE_DIR)/lambda/updater/*.go $(INTERNAL)
	cd $(CODE_DIR) && env GOARCH=amd64 GOOS=linux go build -v $(BUILD_OPT) -o $(CWD)/build/updater $(CODE_DIR)/lambda/updater && cd $(CWD)
$(CWD)/build/matcher: $(CODE_DIR)/lambda/matcher/*.go $(INTERNAL)
	cd $(CODE_DIR) && env GOARCH=amd64 GOOS=linux go build -v $(BUILD_OPT) -o $(CWD)/build/matcher $(CODE_DIR)/lambda/matcher && cd $(CWD)

$(TEMPLATE_FILE): $(SAM_CONFIG) $(BASE_FILE)
	jsonnet -J $(CODE_DIR) $(SAM_CONFIG) -o $(TEMPLATE_FILE)

$(SAM_FILE): $(TEMPLATE_FILE) $(BINPATH) $(STACK_CONFIG)
	aws cloudformation package \
		--region $(shell jsonnet $(STACK_CONFIG) | jq .Region) \
		--template-file $(TEMPLATE_FILE) \
		--s3-bucket $(shell jsonnet $(STACK_CONFIG) | jq .CodeS3Bucket) \
		--s3-prefix $(shell jsonnet $(STACK_CONFIG) | jq .CodeS3Prefix) \
		--output-template-file $(SAM_FILE)

$(OUTPUT_FILE): $(SAM_FILE)
	aws cloudformation deploy \
		--region $(shell jsonnet $(STACK_CONFIG) | jq .Region) \
		--template-file $(SAM_FILE) \
		--stack-name $(STACK_NAME) \
		--capabilities CAPABILITY_IAM \
		$(TAGOPT) \
		--no-fail-on-empty-changeset
	aws cloudformation describe-stack-resources \
		--region $(shell jsonnet $(STACK_CONFIG) | jq .Region) \
		--stack-name $(STACK_NAME) > $(OUTPUT_FILE)

delete:
	aws cloudformation delete-stack \
		--region $(shell jsonnet $(STACK_CONFIG) | jq .Region) \
		--stack-name $(STACK_NAME)
	rm -f $(OUTPUT_FILE)
