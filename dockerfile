FROM golang:1.14

ENV GO111MODULE=on
ENV GOFLAGS=-mod=vendor
ENV APP_USER app
ENV APP_HOME /go/src/reporting-app

# setting working directory
WORKDIR /go/src/app

# installing dependencies
RUN go mod vendor

COPY / /go/src/app/
RUN go build -o reporting-app

EXPOSE 8010

CMD ["./reporting-app"]