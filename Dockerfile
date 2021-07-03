# Stage 1: Build the application
FROM golang:1.16 as builder

RUN mkdir /build

WORKDIR /redirect-server

ADD ./go.mod ./go.sum ./
RUN go mod download

ADD . ./
RUN go build -v -o /build/redirect-server .

# Stage 2: Copy files and configure what we need
FROM debian:buster-slim

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

# Copy the built seabird into the container
COPY --from=builder /build /bin

ENV REDIRECTS_CONFIG /srv/redirect-server/redirects.json

EXPOSE 8080

ENTRYPOINT ["/bin/redirect-server"]
