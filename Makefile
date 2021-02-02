build:
	docker build -t symops/rabbit-amazon-forwarder -f Dockerfile .

push: test build
	docker push symops/rabbit-amazon-forwarder

test:
	docker-compose run --rm tests

dev:
	go build
