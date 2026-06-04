FROM golang:1.22-alpine AS builder

WORKDIR /app

ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=sum.golang.google.cn

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/server ./cmd/server

FROM alpine:3.20

WORKDIR /app

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk add --no-cache ca-certificates

COPY --from=builder /app/server /app/server

ENV APP_ENV=production
ENV APP_PORT=8080
ENV REQUIRE_PERSISTENT_LEDGER=true
ENV WS_ALLOWED_ORIGIN=*

EXPOSE 8080

CMD ["/app/server"]
