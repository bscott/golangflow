test:
	go test
build:
	buffalo build
deploy:
	heroku container:push web
	heroku run ./bin/app migrate