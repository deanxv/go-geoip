<div align="center">

# go-geoip

_基于**MaxMind**的GeoIP库的IP信息查询服务_

[✨演示站](https://iplookup.pro)

[✨演示接口](https://api8.iplookup.pro/ip)

[✨演示接口文档](https://api8.iplookup.pro/swagger/index.html)
</div>


## 功能

- [x] 获取本机或指定IP所在的**IP段**、**ASN**、**城市**、**经度**、**纬度**、**子区域**、**省市区**、**注册国家**。
- [x] 定期(每周)更新GeoLite2库。
- [x] 支持自定义City.mmdb远程地址。

### 接口文档:

`http://<ip>:<port>/swagger/index.html`

### 示例:

<span><img src="docs/img.png" width="800"/></span>

## 如何使用

1. 部署后访问`http://<ip>:<port>/swagger/index.html`查看接口文档。[可选]
2. 使用`/ip`接口查询IP信息。例如：`http://<ip>:<port>/ip`
3. 使用`/ip/{ip}`接口查询指定IP信息。例如：`http://<ip>:<port>/ip/8.8.8.8`

### 基于 Docker-Compose(All In One) 进行部署

```shell
docker-compose pull && docker-compose up -d
```

#### docker-compose.yml

```docker
version: '3.4'

services:
  go-geoip:
    image: deanxv/go-geoip:latest
    container_name: go-geoip
    restart: always
    ports:
      - "7099:7099"
    volumes:
      - ./data:/app/go-geoip/data
    environment:
      - API_SECRET=123456  # [可选]修改此行为请求头校验的值（前后端统一）
      - TZ=Asia/Shanghai
```

### 基于 Docker 进行部署

```docker
docker run --name go-geoip -d --restart always \
-p 7099:7099 \
-v $(pwd)/data:/app/go-geoip/data \
-e API_SECRET="123456" \
-e TZ=Asia/Shanghai \
deanxv/go-geoip
```

其中`API_SECRET`修改为自己的。

如果上面的镜像无法拉取,可以尝试使用 GitHub 的 Docker 镜像,将上面的`deanxv/go-geoip`替换为`ghcr.io/deanxv/go-geoip`即可。

### 部署到第三方平台

<details>
<summary><strong>部署到 Zeabur</strong></summary>
<div>

> Zeabur 的服务器在国外,自动解决了网络的问题,~~同时免费的额度也足够个人使用~~

点击一键部署:

[![Deploy on Zeabur](https://zeabur.com/button.svg)](https://zeabur.com/templates/3KXDY6?referralCode=deanxv)

**一键部署后 `API_SECRET`变量也需要替换！**

或手动部署:

1. 首先 **fork** 一份代码。
2. 进入 [Zeabur](https://zeabur.com?referralCode=deanxv),使用github登录,进入控制台。
3. 在 Service -> Add Service,选择 Git（第一次使用需要先授权）,选择你 fork 的仓库。
4. Deploy 会自动开始,先取消。
5. 添加环境变量

   `PORT=7099` [可选]服务端口
   `API_SECRET=123456` [可选]接口密钥-修改此行为请求头校验的值(多个请以,分隔)
   `TZ=Asia/Shanghai`

保存。

6. 选择 Redeploy。

</div>


</details>

<details>
<summary><strong>部署到 Render</strong></summary>
<div>

> Render 提供免费额度,绑卡后可以进一步提升额度

Render 可以直接部署 docker 镜像,不需要 fork 仓库：[Render](https://dashboard.render.com)

</div>
</details>

## 配置

### 环境变量

1. `PORT=7099`  [可选]服务端口
1. `API_SECRET=123456`  [可选]接口密钥-修改此行为请求头校验的值(多个请以,分隔)(请求header中增加 Authorization:Bearer 123456)
2. `CITY_DB_REMOTE_URL=https://xxx.com/GeoIP2-City.mmdb`  [可选]city.mmdb远程地址
3. `ASN_DB_REMOTE_URL=https://xxx.com/GeoLite2-ASN.mmdb`  [可选]ASN.mmdb远程地址
4. `CN_DB_REMOTE_URL=https://xxx.com/GeoCN.mmdb`  [可选]CN.mmdb远程地址