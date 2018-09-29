.PHONY: all backend

all: backend

backend:
	go build

run: backend
	goreman -f Procfile.dev start

