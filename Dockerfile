FROM golang:1.22 as builder

WORKDIR /go/src/github.com/ludydoo/poc-cloud-service
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /poc-cloud-service
FROM alpine:3.9
COPY --from=builder /poc-cloud-service /app
CMD ["/app"]