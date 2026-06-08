# Sub2API 测试环境（与生产隔离）

在**同一台服务器**上跑一套与生产完全隔离的测试栈。隔离点：独立项目名、独立容器名（`-test` 后缀）、独立端口（默认 8081）、独立网络、独立数据目录、独立密码/密钥。

## 部署步骤

```bash
# 1. 进入测试部署目录（可整个目录上传到服务器任意位置，例如 ~/sub2api-test）
cd deploy-test

# 2. 准备 .env
cp .env.test.example .env
echo "JWT_SECRET=$(openssl rand -hex 32)" >> .env
echo "TOTP_ENCRYPTION_KEY=$(openssl rand -hex 32)" >> .env
# 编辑 .env：设置 POSTGRES_PASSWORD（与生产不同）、ADMIN 等
nano .env

# 3. 建数据目录
mkdir -p data postgres_data redis_data

# 4. 用独立项目名启动（关键！-p 让网络等资源带 sub2api-test 前缀）
docker compose -p sub2api-test up -d

# 5. 查看日志 / 自动生成的管理员密码
docker compose -p sub2api-test logs -f sub2api
docker compose -p sub2api-test logs sub2api | grep "admin password"

# 6. 访问 http://<服务器IP>:8081
```

## 常用命令

```bash
docker compose -p sub2api-test ps          # 查看状态
docker compose -p sub2api-test restart sub2api
docker compose -p sub2api-test pull && docker compose -p sub2api-test up -d   # 升级镜像
docker compose -p sub2api-test down        # 停止（保留数据）
docker compose -p sub2api-test down && rm -rf data postgres_data redis_data   # 彻底清空
```

## 注意事项

- **务必带 `-p sub2api-test`**，否则会和生产的 Compose 项目混在一起。
- 阿里云安全组要放行 `8081` 端口（仅对你信任的 IP 开放更安全）。
- 测试环境一定用**独立的数据库密码和密钥**，不要复用生产的。
- 同机共享 CPU/内存，跑压测前注意给生产留足资源，必要时给容器加 `deploy.resources.limits`。
