FROM golang:1.20-alpine as build

RUN mkdir /usr/app/
WORKDIR /usr/app/
COPY . .

# Unit tests
RUN apk add build-base && go test -buildvcs=false ./...

RUN GOOS=linux GOARCH=amd64 go build -buildvcs=false -ldflags="-w -s" -o /go/bin/mangatsu-server github.com/Mangatsu/server/cmd/mangatsu-server

FROM alpine

RUN adduser -D mangatsu && mkdir /home/mangatsu/app && mkdir /home/mangatsu/data
USER mangatsu
WORKDIR /home/mangatsu/app

COPY --from=build /go/bin/mangatsu-server /home/mangatsu/app/mangatsu-server
COPY --from=build /usr/app/pkg/db/migrations /home/mangatsu/app/pkg/db/migrations

EXPOSE 5050
CMD [ "sh", "-c", "/home/mangatsu/app/mangatsu-server" ]
