test:
	go test
build:
	buffalo build
deploy:
	heroku container:login
	heroku container:push web --app golangflow
	heroku run ./bin/app migrate --app golangflow