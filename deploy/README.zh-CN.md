# 长期维护部署方案

这套目录专门给 `Cloudflare for SaaS + CF 优选 IP + Caddy + Docker` 的生产环境使用。

原则只有三条：

1. 本地开发时改 `E:\new-api`，不要在服务器上编译源码。
2. 发布时使用你自己的镜像，不再直接使用 `calciumion/new-api:latest`。
3. 服务器只负责 `pull + up -d`，入口层继续交给 `Cloudflare + Caddy`。

## 推荐架构

生产链路：

`api.sisyphusx.com -> DNSPod -> Cloudflare -> origin.ymsunv.com -> Caddy -> 127.0.0.1:3000 -> new-api`

预发链路：

`beta-api.sisyphusx.com -> Cloudflare -> beta-origin.ymsunv.com -> Caddy -> 127.0.0.1:3001 -> new-api preview`

边界建议：

- `new-api` 只监听 `127.0.0.1`
- MySQL 不再暴露公网 `3306`
- Caddy 作为唯一公网入口
- Cloudflare for SaaS 和 CF 优选 IP 只管入口，不参与发版

## 目录说明

- `compose.prod.yml`: 生产环境 compose
- `compose.preview.yml`: 预发环境 compose
- `.env.prod.example`: 生产环境变量模板
- `.env.preview.example`: 预发环境变量模板
- `build-and-push.ps1`: 本地构建并推送你自己的镜像
- `Caddyfile.origin.example`: 适合当前 origin 架构的 Caddy 示例

## 本地开发

前端本地热更新：

```powershell
cd E:\new-api\web
bun run dev
```

后端本地运行：

```powershell
cd E:\new-api
go run .
```

Vite 已经代理 `/api`、`/mj`、`/pg` 到 `http://localhost:3000`。

## 首次生产部署

服务器上建议准备：

```bash
mkdir -p /data/one-api
mkdir -p /data/one-api/mysql
cp deploy/.env.prod.example deploy/.env.prod
```

然后编辑 `deploy/.env.prod`：

- `NEW_API_IMAGE` 改成你自己的镜像
- `MYSQL_ROOT_PASSWORD` 改成真实密码
- `SESSION_SECRET` 改成随机长字符串
- `FRONTEND_BASE_URL` 改成真实入口域名

启动生产环境：

```bash
docker-compose --env-file deploy/.env.prod -f deploy/compose.prod.yml up -d
```

## 首次预发部署

服务器上：

```bash
mkdir -p /data/new-api-preview
mkdir -p /data/new-api-preview/mysql
cp deploy/.env.preview.example deploy/.env.preview
```

然后编辑 `deploy/.env.preview`。

启动预发环境：

```bash
docker-compose --env-file deploy/.env.preview -f deploy/compose.preview.yml up -d
```

## 发版方式一：本地直接构建并推送

先登录镜像仓库。例如 GHCR：

```powershell
echo $env:GITHUB_TOKEN | docker login ghcr.io -u <github-user> --password-stdin
```

构建并推送：

```powershell
cd E:\new-api
.\deploy\build-and-push.ps1 -ImageRepository ghcr.io/<owner>/<repo> -AlsoTagLatest
```

脚本会：

- 自动生成带时间戳和 git short sha 的 tag
- 暂时写入 `VERSION`
- 构建镜像
- 推送镜像
- 打印服务器更新命令

## 发版方式二：GitHub Actions 构建镜像

仓库里新增了 `publish-custom-image.yml` 工作流。

前提：

- 这条链路需要你把仓库 fork 到自己的 GitHub 账号，或者把当前仓库推到你自己可控的仓库
- 如果 `origin` 仍然指向上游 `QuantumNous/new-api`，你只能先用本地 `build-and-push.ps1`

使用方式：

1. 推送代码到你自己的 GitHub 仓库
2. 在 Actions 里手动运行 `Publish Custom Image`
3. 默认会发布到 `ghcr.io/<owner>/<repo>:<tag>`

然后服务器拉取新镜像：

```bash
docker-compose --env-file deploy/.env.prod -f deploy/compose.prod.yml pull new-api
docker-compose --env-file deploy/.env.prod -f deploy/compose.prod.yml up -d new-api
```

## Caddy 建议

把生产和预发都只转发到本机端口：

- 生产：`127.0.0.1:3000`
- 预发：`127.0.0.1:3001`

示例见 `Caddyfile.origin.example`。

注意：

- 流式接口需要尽量少缓冲，所以示例里使用了 `flush_interval -1`
- Cloudflare 真实 IP 的 trusted proxies 建议在你的真实 Caddyfile 里单独维护，不放进仓库

## 迁移你现在的服务器

你当前服务器已经在用：

- `/data/one-api`
- `/data/one-api/mysql`
- Caddy
- `api.sisyphusx.com` 作为 Cloudflare SaaS 入口

所以迁移成本很低：

1. 把当前 `new-api` 切换成 `127.0.0.1:3000`
2. 把 MySQL 改成只走容器内网
3. 把 compose 切换到 `deploy/compose.prod.yml`
4. 以后只更新 `NEW_API_IMAGE`

## 最小发版命令

如果镜像已经推送完成，服务器上实际只需要：

```bash
docker-compose --env-file deploy/.env.prod -f deploy/compose.prod.yml pull new-api
docker-compose --env-file deploy/.env.prod -f deploy/compose.prod.yml up -d new-api
```
