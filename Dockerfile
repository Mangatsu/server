FROM golang:1.17-alpine as build

RUN mkdir /usr/app/
WORKDIR /usr/app/
COPY . .

# Unit tests
RUN apk add build-base && go test ./...

RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go/bin/mangatsu-server github.com/Mangatsu/server/cmd/mangatsu-server

FROM alpine

RUN adduser -D mangatsu && mkdir /home/mangatsu/app && mkdir /home/mangatsu/data
USER mangatsu
WORKDIR /home/mangatsu/app

COPY --from=build /go/bin/mangatsu-server /home/mangatsu/app/mangatsu-server

EXPOSE 5000
CMD [ "sh", "-c", "/home/mangatsu/app/mangatsu-server" ]
