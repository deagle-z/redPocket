# ---------- 构建阶段 ----------
FROM golang:1.23.5-bullseye AS builder
WORKDIR /app
COPY go.mod go.sum ./
COPY vendor ./vendor
COPY app ./app
COPY core ./core
COPY main.go ./
COPY *.yaml ./
COPY dist ./dist
# 安装 swag
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN swag init -g main.go
# 开启 CGO，指定 GOOS、GOARCH，使用 vendor 模式
ENV CGO_ENABLED=1 GOOS=linux GOARCH=amd64 GOFLAGS="-mod=vendor"
RUN go build -o main
# ---------- 运行阶段 ----------
FROM debian:bookworm
RUN apt-get update && apt-get install -y --no-install-recommends tzdata ca-certificates && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/*.yaml ./
COPY --from=builder /app/dist ./dist
COPY --from=builder /app/core/locales ./core/locales
COPY --from=builder /app/docs ./docs
ENV TZ=Asia/Shanghai
EXPOSE 8080
ENTRYPOINT ["/app/main"]