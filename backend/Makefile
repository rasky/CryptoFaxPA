.PHONY: all backend

all: backend

backend:
	go build -o cryptofax-backend

run: backend
	goreman -f Procfile.dev start

