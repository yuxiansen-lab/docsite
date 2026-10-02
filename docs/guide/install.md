# 安装

## 系统要求

- Go 1.22 及以上
- 目标平台：Linux / macOS / Windows

## 构建

```bash
go build -o docsite ./cmd/server
```

## 运行

```bash
./docsite -addr :8080
```

## 从源码运行（开发）

```bash
go run ./cmd/server -addr :3000
```

- 开发时可直接用任意静态服务器托管 `web/`，API 指向本服务
- 生产建议直接运行单二进制

## 任务清单

- [x] 验证 `GET /api/config` 返回站点配置
- [x] 验证 `GET /api/menu` 返回菜单树
- [ ] 配置反向代理（可选）
