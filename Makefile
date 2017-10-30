build:
	docker build -t airhelp/rabbit-amazon-forwarder -f Dockerfile .

push: test build
	docker push airhelp/rabbit-amazon-forwarder

test:
	docker-compose build --pull
	docker-compose run --rm tests

dev:
	go build
