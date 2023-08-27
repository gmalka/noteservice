FROM golang:1.19 AS builder
WORKDIR /noteservice
COPY . .
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./app ./cmd/main.go

FROM alpine:latest
WORKDIR /noteservice
COPY ./.env .
COPY ./public ./public
COPY ./templates ./templates
COPY --from=builder /noteservice/app .
ENTRYPOINT [ "/noteservice/app" ]