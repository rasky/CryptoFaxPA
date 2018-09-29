.PHONY: all backend

all: backend

.env:
	heroku config --shell >.env

backend:
	go build

run: .env backend
	goreman -f Procfile.dev start
