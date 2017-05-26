FROM gobuffalo/buffalo:latest

RUN mkdir -p $GOPATH/src/github.com/bscott/golangflow
WORKDIR $GOPATH/src/github.com/bscott/golangflow
ADD . .
RUN npm install
RUN buffalo build -o bin/app
CMD ./bin/app