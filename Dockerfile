FROM golang:1.22 as builder

WORKDIR /go/src/github.com/ludydoo/poc-cloud-service
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN GOOS=linux CGO_ENABLED=0 go build -o app .
FROM alpine:latest
COPY --chown=65534:65534 --from=builder /go/src/github.com/ludydoo/poc-cloud-service/app .
USER 65534:65534
CMD ["./app"]
