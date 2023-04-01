FROM golang:1.19.6 AS build_base

RUN apt-get install -y git
# Set the Current Working Directory inside the container
WORKDIR /tmp/yaam2
# We want to populate the module cache based on the go.{mod,sum} files.
COPY . .
RUN go mod download
# Unit tests
RUN go test ./... -v
# Build the Go app
RUN CGO_ENABLED=0 go build -o ./out/yaam2 main.go

# Start fresh from a smaller image
FROM alpine:3.9 

COPY --from=build_base /tmp/yaam2/out/yaam2 /app/
# COPY example.yaml /app/yaam2.yaml
# This container exposes port 3000 to the outside world
EXPOSE 25213
# Run the binary program produced by `go install`
WORKDIR /app
CMD ["yaam2"]