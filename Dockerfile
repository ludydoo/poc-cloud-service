# Building the UI
FROM node:22 as ui-builder
WORKDIR /app
COPY ui/package.json ui/pnpm-lock.yaml ./
RUN npm install -g pnpm && pnpm install
COPY ui .
RUN pnpm build

# Building the Go binary
FROM golang:1.22 as builder
WORKDIR /go/src/github.com/ludydoo/poc-cloud-service
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN GOOS=linux CGO_ENABLED=0 go build -o app .

# Building the final image
FROM alpine:latest
WORKDIR /app
COPY --chown=65534:65534 --from=builder /go/src/github.com/ludydoo/poc-cloud-service/app .
COPY --chown=65534:65534 --from=ui-builder /app/dist ./ui/dist
USER 65534:65534
ENTRYPOINT ["./app"]
CMD ["serve"]
