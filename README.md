<div align="center">

# go-geoip


</div>

### 接口文档:

`http://<ip>:<port>/swagger/index.html`


### 示例:

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

> Zeabur 的服务器在国外,自动解决了网络的问题,同时免费的额度也足够个人使用

点击一键部署:


**一键部署后 `API_SECRET`变量也需要替换！**

或手动部署:

1. 首先 **fork** 一份代码。
2. 进入 [Zeabur](https://zeabur.com?referralCode=deanxv),使用github登录,进入控制台。
3. 在 Service -> Add Service,选择 Git（第一次使用需要先授权）,选择你 fork 的仓库。
4. Deploy 会自动开始,先取消。
5. 添加环境变量

   `API_SECRET:123456` [可选]接口密钥-修改此行为请求头校验的值(多个请以,分隔)

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

1. `API_SECRET=123456`  [可选]接口密钥-修改此行为请求头校验的值(多个请以,分隔)