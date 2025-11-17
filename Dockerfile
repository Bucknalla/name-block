FROM golang:1.21-alpine AS build

WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download

COPY main.go .
COPY data ./data

RUN CGO_ENABLED=0 GOOS=linux go build -o name-block main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=build /app/name-block .
COPY --from=build /app/data ./data

CMD ["./name-block"]
