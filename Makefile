IMAGE ?= ats-jammy

.PHONY: build check test run exec

build:
	mkdir -p ./log
	docker build $(DOCKER_BUILD_OPTS) -t $(IMAGE) . 2>&1 | tee ./log/build.log

test:
	mkdir -p ./log
	docker run --rm -t --entrypoint=/usr/local/go/bin/go $(IMAGE) test -v -debug 2>&1 \
		| tee ./log/test.log

build_and_test:
	echo IMAGE=$(IMAGE) GITHUB_REPO=$(GITHUB_REPO) GITHUB_BRANCH=$(GITHUB_BRANCH)
	mkdir -p ./log
	docker build $(DOCKER_BUILD_OPTS) --build-arg GITHUB_REPO=$(GITHUB_REPO) --build-arg GITHUB_BRANCH=$(GITHUB_BRANCH) -t $(IMAGE) . 2>&1 | tee ./log/$(IMAGE)-build.log
	docker run --rm -t --entrypoint=/usr/local/go/bin/go $(IMAGE) test -v -debug 2>&1 | tee ./log/$(IMAGE)-test.log

build_and_test_ats921:
	$(MAKE) build_and_test IMAGE=ats921 GITHUB_REPO=https://github.com/apache/trafficserver GITHUB_BRANCH=9.2.1

build_and_test_ats921fix:
	$(MAKE) build_and_test IMAGE=ats921fix GITHUB_REPO=https://github.com/hnakamur/trafficserver GITHUB_BRANCH=9_2_1_dont_add_content_length_for_status_204_cache

build_and_test_ats_master:
	$(MAKE) build_and_test IMAGE=atsmaster GITHUB_REPO=https://github.com/apache/trafficserver GITHUB_BRANCH=master

build_and_test_ats_master_fix:
	$(MAKE) build_and_test IMAGE=atsmasterfix GITHUB_REPO=https://github.com/hnakamur/trafficserver GITHUB_BRANCH=dont_add_content_length_for_status_204_cache

build_and_test_all: build_and_test_ats921 build_and_test_ats921fix build_and_test_ats_master build_and_test_ats_master_fix
