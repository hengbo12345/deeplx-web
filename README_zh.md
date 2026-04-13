# DeepLX Web

一个现代化的 DeepLX Web 界面 - 提供免费的文本和 Word 文档翻译服务。

## 功能特性

- **文本翻译**：支持 28+ 种语言互译，具有自动检测功能
- **文档翻译**：上传并翻译 Word 文档（.docx），保留原有格式
- **现代化界面**：基于 React 和 Tailwind CSS 构建的简洁响应式界面
- **文件上传**：支持最大 10MB 的文档
- **实时翻译**：由 DeepLX 驱动的高速翻译
- **免费开源**：无需 API 密钥，完全免费使用

## 支持的语言

保加利亚语、中文、捷克语、丹麦语、荷兰语、英语、爱沙尼亚语、芬兰语、法语、德语、希腊语、匈牙利语、印尼语、意大利语、日语、韩语、立陶宛语、拉脱维亚语、挪威语、波兰语、葡萄牙语、罗马尼亚语、俄语、斯洛伐克语、斯洛文尼亚语、西班牙语、瑞典语、土耳其语、乌克兰语

## 技术栈

### 后端
- Go 1.21
- Chi 路由器
- Zap 日志
- Lumberjack 日志轮转
- DOCX 文档处理

### 前端
- React 18
- TypeScript
- Vite
- Tailwind CSS
- React Router

## 项目结构

```
.
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go          # 服务入口
│   ├── internal/
│   │   ├── config/              # 配置管理
│   │   ├── handler/             # HTTP 处理器
│   │   ├── middleware/          # HTTP 中间件
│   │   ├── models/              # 数据模型
│   │   └── service/             # 业务逻辑
│   ├── pkg/
│   │   └── utils/               # 工具函数
│   ├── go.mod
│   ├── go.sum
│   └── Dockerfile
├── frontend/
│   ├── src/
│   │   ├── components/          # React 组件
│   │   ├── pages/               # 页面组件
│   │   ├── lib/                 # 工具和 API 客户端
│   │   └── styles/              # 全局样式
│   ├── package.json
│   ├── vite.config.ts
│   └── tailwind.config.js
├── docker-compose.yml
└── README.md
```

## 快速开始

### 前置要求

- Go 1.21 或更高版本
- Node.js 18 或更高版本
- Docker 和 Docker Compose（可选）

### 部署方式

#### 方式一：Docker Compose 部署（开发/测试环境）

1. 克隆仓库：
```bash
git clone <repository-url>
cd sayHello
```

2. 配置环境变量：
```bash
cp .env.example .env
# 编辑 .env 填入 DEEPLX_TOKEN
```

3. 启动服务：
```bash
docker-compose up -d
```

4. 访问应用：
- 前端：http://localhost:9449
- 后端 API：http://localhost:8449

#### 方式二：生产环境 HTTPS 部署

本项目已配置为支持 HTTPS 生产环境部署（域名：deeplxweb.yourdoamin.com）。

1. 准备 SSL 证书：
```bash
# 将 SSL 证书文件放到 frontend/ssl/ 目录
cp your-cert.cer frontend/ssl/deeplxweb.cer
cp your-cert.key frontend/ssl/deeplxweb.key
```

2. 配置环境变量：
```bash
cp .env.example .env
# 编辑 .env 填入 DEEPLX_TOKEN
```

3. 修改 nginx.conf 中的域名（如需要）：
```nginx
server_name  your-domain.com;  # 修改为你的域名
```

4. 构建并启动：
```bash
docker-compose up -d
```

5. 访问：https://deeplxweb.yourdoamin.com（或你配置的域名）

**注意**：生产环境配置中，Nginx 会代理请求到本地的 127.0.0.1:8449 端口，请确保后端服务在该端口运行。

### 手动安装

#### 后端

1. 进入后端目录：
```bash
cd backend
```

2. 安装依赖：
```bash
go mod download
```

3. 创建 `.env` 文件（可选）：
```env
SERVER_PORT=9448
DEEPLX_URL=http://localhost:1188
DEEPLX_TOKEN=
UPLOAD_PATH=./uploads
FILE_MAX_AGE=24h
UPLOAD_MAX_SIZE=10485760
CLEANUP_INTERVAL=1h
LOG_LEVEL=info
```

4. 运行服务：
```bash
go run cmd/server/main.go
```

后端默认运行在 9448 端口。

#### 前端

1. 进入前端目录：
```bash
cd frontend
```

2. 安装依赖：
```bash
npm install
```

3. 启动开发服务器：
```bash
npm run dev
```

前端默认运行在 3000 端口。

### 生产构建

#### 后端
```bash
cd backend
go build -o deeplx-web ./cmd/server
```

#### 前端
```bash
cd frontend
npm run build
```

构建文件将输出到 `frontend/dist` 目录。

## API 端点

### 健康检查
```
GET /health
```

### 文本翻译
```
POST /api/translate
Content-Type: application/json

{
  "text": "你好，世界！",
  "source_lang": "ZH",
  "target_lang": "EN"
}
```

### 文档翻译
```
POST /api/translate/document
Content-Type: multipart/form-data

file: <文档>
source_lang: ZH
target_lang: EN
```

## 配置

### 后端环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| SERVER_PORT | 服务端口 | 9448 |
| DEEPLX_URL | DeepLX 服务地址 | http://localhost:1188 |
| DEEPLX_TOKEN | DeepLX 认证 Token | (空) |
| UPLOAD_PATH | 文件上传目录 | ./uploads |
| FILE_MAX_AGE | 文件保留时间 | 24h |
| UPLOAD_MAX_SIZE | 最大文件大小 | 10485760 (10MB) |
| CLEANUP_INTERVAL | 清理间隔 | 1h |
| LOG_LEVEL | 日志级别 | info |

## 开发

### 后端开发
```bash
cd backend
go run cmd/server/main.go
```

### 前端开发
```bash
cd frontend
npm run dev
```

### 代码检查

前端：
```bash
npm run lint
```

## Docker

### 构建后端镜像
```bash
docker build -t deeplx-web ./backend
```

### 使用 Docker 运行
```bash
docker run -p 9448:9448 \
  -e DEEPLX_URL=http://your-deeplx-service:1188 \
  deeplx-web
```

## 许可证

本项目是开源的，基于 MIT 许可证。

## 致谢

- [DeepLX](https://github.com/OwO-Network/DeepLX) - 非官方的开源 DeepL 翻译 API
- [Go Chi](https://github.com/go-chi/chi) - 轻量级、符合 Go 习惯的可组合 HTTP 服务路由器
- [React](https://react.dev/) - 用于构建用户界面的 JavaScript 库
- [Tailwind CSS](https://tailwindcss.com/) - 实用优先的 CSS 框架

## 支持

如有问题、疑问或贡献，请访问项目仓库。
