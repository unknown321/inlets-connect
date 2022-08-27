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
