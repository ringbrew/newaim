FROM dhub.kubesre.xyz/golang:1.19-alpine as basic

ENV GOPROXY="https://goproxy.cn"

WORKDIR /app

ADD . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o productsearch cmd/productsearch.go

FROM dhub.kubesre.xyz/alpine:3.12

WORKDIR /app

COPY --from=basic /app/productsearch /app/productsearch

ENTRYPOINT ["/app/productsearch"]
EXPOSE 3545
