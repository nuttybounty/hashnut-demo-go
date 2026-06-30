# HashNut 示范商城 (Go)

演示如何使用 [hashnut-sdk-go](https://github.com/nuttybounty/hashnut-sdk-go) (v4) 对接 HashNut 支付 API（V4 版本）的示范商户应用。支持多链支付（ERC20 / TRC20）。

## 技术栈

- **后端**: Go + Gin
- **数据库**: PostgreSQL
- **支付**: [HashNut Go SDK](https://github.com/nuttybounty/hashnut-sdk-go) (V4)
- **前端**: [hashnut-demo-web](https://github.com/nuttybounty/hashnut-demo-web) (React + TypeScript)

## 环境要求

- Go 1.21+
- PostgreSQL 14+
- Node.js 18+（前端）
- HashNut 商户账号及 API Key
- ngrok（本地开发回调用）

## 快速开始

### 1. 克隆项目

```bash
git clone https://github.com/nuttybounty/hashnut-demo-go.git
cd hashnut-demo-go
```

### 2. 创建数据库

```bash
psql -U postgres -c "CREATE DATABASE demo_shop;"
```

### 3. 配置种子数据

编辑 `migrate.sql`，填入你的 API 密钥：

```sql
INSERT INTO t_hashnut_api_key (chain_code, splitter, access_key_id, secret_key) VALUES
    ('erc20', '0x...你的ETH分账合约地址...', '你的access-key-id', '你的secret-key'),
    ('trc20', 'T...你的Tron分账合约地址',    '你的access-key-id', '你的secret-key');
```

然后执行：

```bash
psql -U postgres -d demo_shop -f migrate.sql
```

### 4. 配置

编辑 `etc/application.yaml`：

```yaml
server:
  port: 1800

database:
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "postgres"
  dbname: "demo_shop"
  sslmode: "disable"

hashnut:
  testMode: false    # true = 测试网, false = 正式环境
  baseURL: ""        # 留空使用默认地址
```

### 5. 启动后端

```bash
go run main.go
```

服务启动在 `http://localhost:1800`。

### 6. 启动前端

```bash
git clone https://github.com/nuttybounty/hashnut-demo-web.git
cd hashnut-demo-web
npm install
npm run dev
```

前端启动在 `http://localhost:5173`，自动代理 `/api` 到 `localhost:1800`。

打开浏览器访问 `http://localhost:5173`。

## 本地开发 ngrok 配置

HashNut 后端需要向你的本地机器发送支付通知，需要用 ngrok 暴露端口：

### 1. 启动 ngrok

```bash
ngrok http 1800
```

会得到一个公网 URL，例如 `https://xxxxx.ngrok-free.dev`。

### 2. 配置 HashNut API Key

在 HashNut 商户后台设置：

| 字段 | 本地开发 | 正式生产 |
|------|---------|---------|
| notifyURL | `https://xxxxx.ngrok-free.dev/api/notify` | `https://你的域名/api/notify` |
| callbackURL | `http://localhost:5173/payment-result` | `https://你的域名/payment-result` |

- **notifyURL**: 后端回调通知 — 必须公网可达，使用 ngrok URL
- **callbackURL**: 前端支付完成跳转 — 本地开发用 `localhost:5173`（浏览器跳转，不需要公网）

### 3. 本地支付流程

```
浏览器 (localhost:5173)
  -> 点击 "Pay with Crypto"
  -> POST /api/orders（Vite 代理到 localhost:1800）
  -> 跳转到 HashNut 支付页面 (defi.hashnut.io/pay)
  -> 用户链上支付
  -> HashNut 后端发送通知到 ngrok URL -> localhost:1800/api/notify
  -> HashNut 前端跳转到 http://localhost:5173/payment-result?state=4&...
  -> 前端显示支付成功
```

## 接口说明

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/products` | 获取商品列表 |
| GET | `/api/chains` | 获取支持的链+币种（从数据库读取） |
| POST | `/api/orders` | 创建订单 `{productId, chainCode, coinCode}` |
| GET | `/api/orders/:id` | 查询订单状态 |
| POST | `/api/orders/:id/confirm` | 提交支付交易哈希 `{payTxId}` |
| POST | `/api/notify` | HashNut 支付结果回调 |

## 数据库表结构

| 表 | 说明 |
|----|------|
| `t_coin_info` | 支持的链+币种配置 |
| `t_hashnut_api_key` | 每条链的 splitter 地址 + API 密钥 |
| `products` | 商品（只有价格，不绑定链/币种） |
| `orders` | 订单（记录用户选择的链+币种） |

## 许可证

MIT
