IMAGE_NAME ?= claude-agent
IMAGE_TAG  ?= latest
REGISTRY   ?= ghcr.io/adeotek

.PHONY: build push run clean

build:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .

push: build
	docker tag $(IMAGE_NAME):$(IMAGE_TAG) $(REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)
	docker push $(REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)

run:
	@[ -n "$(REPO_URL)" ] || (echo "Usage: make run REPO_URL=https://github.com/org/repo TASK=\"your task\""; exit 1)
	docker run --rm \
	  -v ~/.claude.json:/root/.claude.json:ro \
	  -e GITHUB_TOKEN \
	  -e GIT_AUTHOR_NAME \
	  -e GIT_AUTHOR_EMAIL \
	  -e REPO_URL="$(REPO_URL)" \
	  -v "$(PWD)/output:/output" \
	  $(IMAGE_NAME):$(IMAGE_TAG) "$(TASK)"

clean:
	docker rmi $(IMAGE_NAME):$(IMAGE_TAG) 2>/dev/null || true
