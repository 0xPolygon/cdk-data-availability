# CONTAINER FOR BUILDING BINARY
FROM golang:1.19 AS build

# INSTALL DEPENDENCIES
RUN go install github.com/gobuffalo/packr/v2/packr2@v2.8.3
COPY go.mod go.sum /src/
WORKDIR /src
RUN go mod download

# BUILD BINARY
COPY . /src

WORKDIR /src/db
RUN packr2

WORKDIR /src
RUN make build

# CONTAINER FOR RUNNING BINARY
FROM alpine:3.16.0
COPY --from=build /src/dist/cdk-data-availability /app/cdk-data-availability
EXPOSE 8444
CMD ["/bin/sh", "-c", "/app/cdk-data-availability run"]
