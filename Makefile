TAG = latest
NAME=inlets-connect
IMAGE = $(NAME):$(TAG)
DC := IMAGE=$(IMAGE) docker-compose

build:
	docker build --no-cache -t $(IMAGE) .

up:
	$(DC) up -d

logs:
	$(DC) logs -f

down:
	$(DC) down

lint:
	docker run --rm -t -v $(CURDIR):$(CURDIR) -w $(CURDIR) \
		golangci/golangci-lint:v1.49.0 golangci-lint run
