# 使用 Golang 镜像作为构建阶段
FROM golang AS builder

# 设置环境变量
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux

# 设置工作目录
WORKDIR /build

# 复制 go.mod 和 go.sum 文件,先下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制整个项目并构建可执行文件
COPY . .
RUN go build -o /ip2region-geoip

# 使用 Alpine 镜像作为最终镜像
FROM alpine

# 安装基本的运行时依赖及工具
RUN apk --no-cache add ca-certificates tzdata curl dcron

# 从构建阶段复制可执行文件
COPY --from=builder /ip2region-geoip /app/ip2region-geoip/ip2region-geoip

# 下载初始GeoIP数据库文件
RUN mkdir -p /app/ip2region-geoip/data && cd /app/ip2region-geoip/data \
    && curl -L -o GeoLite2-City.mmdb "https://github.com/P3TERX/GeoLite.mmdb/raw/download/GeoLite2-City.mmdb" \
    && curl -L -o GeoLite2-ASN.mmdb "https://github.com/P3TERX/GeoLite.mmdb/raw/download/GeoLite2-ASN.mmdb" \
    && curl -L -o GeoCN.mmdb "http://github.com/ljxi/GeoCN/releases/download/Latest/GeoCN.mmdb"

# 创建crontab文件
RUN echo "0 0 * * * curl -L -o /app/ip2region-geoip/data/GeoLite2-City.mmdb 'https://github.com/P3TERX/GeoLite.mmdb/raw/download/GeoLite2-City.mmdb' && curl -L -o /app/ip2region-geoip/data/GeoLite2-ASN.mmdb 'https://github.com/P3TERX/GeoLite.mmdb/raw/download/GeoLite2-ASN.mmdb' && curl -L -o /app/ip2region-geoip/data/GeoCN.mmdb 'http://github.com/ljxi/GeoCN/releases/download/Latest/GeoCN.mmdb' && kill -HUP \$(pgrep ip2region-geoip)" > /etc/crontabs/root

# 暴露端口
EXPOSE 7099

# 工作目录
WORKDIR /app/ip2region-geoip

# 设置入口命令，同时启动Cron和应用
CMD ["sh", "-c", "crond && ./ip2region-geoip"]
