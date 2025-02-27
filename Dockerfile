# CONTAINER FOR BUILDING BINARY
FROM golang:1.22 AS build

WORKDIR $GOPATH/src/github.com/0xPolygon/cdk-data-availability

# INSTALL DEPENDENCIES
COPY go.mod go.sum ./
RUN go mod download

# BUILD BINARY
COPY . .
RUN go install github.com/antithesishq/antithesis-sdk-go/tools/antithesis-go-instrumentor@v0.4.3 \
  && mkdir /antithesis \
  && antithesis-go-instrumentor . /antithesis
RUN cd /antithesis/customer && make build

# CONTAINER FOR RUNNING BINARY
FROM debian:bookworm-slim

COPY --from=build /antithesis/customer/dist/cdk-data-availability /app/cdk-data-availability

EXPOSE 8444

CMD ["/bin/sh", "-c", "/app/cdk-data-availability run"]
