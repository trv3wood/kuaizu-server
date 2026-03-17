# 部署配置说明

## GitHub Secrets 配置

在 GitHub 仓库的 Settings > Secrets and variables > Actions 中添加以下 secrets：

### SSH 连接配置
- `SSH_HOST`: 服务器 IP 地址或域名
- `SSH_PORT`: SSH 端口（默认 22，可选）
- `SSH_USER`: SSH 用户名
- `SSH_PRIVATE_KEY`: SSH 私钥内容

### 环境变量配置
- `ENV_DOCKER`: `.env.docker` 文件内容。包含数据库、微信、SMTP、OSS 等所有运行时所需的环境变量。内容格式如下：
  ```
  PORT=8080
  DATABASE_URL=postgres://...
  WECHAT_APPID=...
  WECHAT_SECRET=...
  JWT_SECRET=...
  REGISTER_JWT_SECRET=...
  WECHAT_MCH_ID=...
  WECHAT_MCH_SERIAL_NO=...
  WECHAT_MCH_API_KEY=...
  WECHAT_MCH_PRIVATE_KEY=...
  WECHAT_NOTIFY_URL=...
  WECHAT_PAY_PUBLIC_KEY=...
  WECHAT_PAY_PUBLIC_KEY_ID=...
  SMTP_HOST=...
  SMTP_PORT=...
  SMTP_USER=...
  SMTP_PASSWORD=...
  SMTP_FROM_NAME=快组校园
  BASE_URL=...
  TEST_EMAIL=...
  ADMIN_PORT=8081
  ADMIN_JWT_SECRET=...
  OSS_ACCESS_KEY_ID=...
  OSS_ACCESS_KEY_SECRET=...
  OSS_ENDPOINT=...
  OSS_BUCKET_NAME=...
  OSS_BASE_PATH=uploads
  OSS_DOMAIN=...
  ```

## SSH 密钥生成

在本地生成 SSH 密钥对：

```bash
ssh-keygen -t rsa -b 4096 -C "github-actions" -f ~/.ssh/github_actions
```

将公钥添加到服务器：

```bash
ssh-copy-id -i ~/.ssh/github_actions.pub user@your-server
```

将私钥内容复制到 GitHub Secrets 的 `SSH_PRIVATE_KEY`：

```bash
cat ~/.ssh/github_actions
```

## 触发部署

### 方式 1: 推送 tag
```bash
git tag v1.0.0
git push origin v1.0.0
```

### 方式 2: 手动触发
在 GitHub 仓库的 Actions 页面，选择 "Deploy to Server" workflow，点击 "Run workflow"。

## 部署流程

1. 检出代码
2. 配置 SSH 连接
3. 在服务器上创建 .env.docker 文件
4. 使用 rsync 同步代码到服务器
5. 在服务器上构建 Docker 镜像
6. 重启服务
7. 清理旧镜像

## 服务器要求

- 已安装 Docker 和 Docker Compose
- SSH 访问权限
- 足够的磁盘空间用于构建镜像
