BINARY  := iota
CMD_DIR := ./cmd/server
WORKERS := python javascript shell

.PHONY: all build clean run docker-all $(WORKERS:%=docker-%)

all: build docker-all

build:
	go build -o $(BINARY) $(CMD_DIR)

docker-python:
	docker build -t iota-python ./worker/python

docker-javascript:
	docker build -t iota-javascript ./worker/javascript

docker-shell:
	docker build -t iota-shell ./worker/shell

docker-all: docker-python docker-javascript docker-shell

run: build
	./$(BINARY)

clean:
	rm -f $(BINARY)
	for img in $(WORKERS:%=iota-%); do docker rmi $$img 2>/dev/null || true; done
