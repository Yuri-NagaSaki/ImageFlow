# ImageFlow - 现代图像服务系统

<div align="center">

[中文文档](README_zh.md)
|
[部署说明](https://catcat.blog/imageflow-install.html)
</div>

<div align="center">
  <img src="https://raw.githubusercontent.com/Yuri-NagaSaki/ImageFlow/main/favicon/favicon.svg" alt="ImageFlow Logo" width="120" height="120" style="background: linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%); padding: 20px; border-radius: 16px;">
  <h3>高效智能的图像管理和分发系统</h3>
</div>

ImageFlow 是一个为现代网站和应用程序设计的高效图像服务系统。它能根据设备类型自动提供最合适的图像，并支持 WebP 和 AVIF 等现代图像格式，显著提升网站性能和用户体验。

## ✨ 主要特性

- **API 密钥认证**：安全的 API 密钥验证机制，保护您的图片上传功能
- **自适应图像服务**：根据设备类型（桌面端/移动端）自动提供横向或纵向图片
- **现代格式支持**：自动检测浏览器兼容性并提供 WebP 或 AVIF 格式图片
- **图片过期功能**：支持设置图片过期时间，过期后自动删除（支持本地和S3存储）
- **简单的 API**：通过简单的 API 调用获取随机图片，支持标签过滤
- **用户友好的上传界面**：支持拖拽上传，具有暗黑模式、实时预览和标签管理功能
- **图片管理功能**：通过直观的管理界面查看、筛选和删除图片
- **自动图像处理**：上传后自动检测图像方向并转换为多种格式
- **异步处理**：图像转换在后台进行，不影响主服务
- **高性能**：优化的网络性能以减少加载时间
- **易于部署**：简单的配置和部署流程
- **多存储支持**：支持本地存储和 S3 兼容存储（如 R2）

## 🚀 技术优势

1. **安全性**：API 密钥验证机制确保图片上传和管理功能的安全访问
2. **格式转换**：自动将上传的图片转换为 WebP 和 AVIF 格式，减少 30-50% 的文件大小
3. **设备适配**：为不同设备提供最合适的图片方向
4. **图片生命周期管理**：支持设置图片过期时间，在所有存储类型中自动清理过期图片
5. **热重载**：上传的图片无需重启服务即可立即可用
6. **并发处理**：使用 Go 的并发特性高效处理图像转换
7. **一致性管理**：删除图片时，所有相关格式（原始、WebP、AVIF）会同时被移除
8. **可扩展性**：模块化设计便于扩展和定制
9. **响应式设计**：完美适配桌面端和移动端设备
10. **暗黑模式支持**：自动适应系统主题，支持手动切换
11. **灵活存储**：支持本地和 S3 兼容存储，通过 .env 文件轻松配置

## 📸 界面预览

<div align="center">
  <img src="https://raw.githubusercontent.com/Yuri-NagaSaki/ImageFlow/main/docs/img/imageflow1.png" alt="ImageFlow">
  <img src="https://raw.githubusercontent.com/Yuri-NagaSaki/ImageFlow/main/docs/img/imageflow2.png" alt="ImageFlow">
  <img src="https://raw.githubusercontent.com/Yuri-NagaSaki/ImageFlow/main/docs/img/imageflow3.png" alt="ImageFlow">
  <img src="https://raw.githubusercontent.com/Yuri-NagaSaki/ImageFlow/main/docs/img/imageflow4.png" alt="ImageFlow">
  <img src="https://raw.githubusercontent.com/Yuri-NagaSaki/ImageFlow/main/docs/img/imageflow5.png" alt="ImageFlow">
  <img src="https://raw.githubusercontent.com/Yuri-NagaSaki/ImageFlow/main/docs/img/imageflow6.png" alt="ImageFlow">
  <img src="https://raw.githubusercontent.com/Yuri-NagaSaki/ImageFlow/main/docs/img/imageflow7.png" alt="ImageFlow">
  <img src="https://raw.githubusercontent.com/Yuri-NagaSaki/ImageFlow/main/docs/img/imageflow8.png" alt="ImageFlow">
  <img src="https://raw.githubusercontent.com/Yuri-NagaSaki/ImageFlow/main/docs/img/imageflow9.png" alt="ImageFlow">
</div>

## 🔧 快速开始

### 前置条件

- Go 1.22 或更高版本
- Node.js 18 或更高版本（用于构建前端）
- WebP 工具（`libwebp-tools`）
- AVIF 工具（`libavif-apps`）
- Docker 和 Docker Compose（可选，用于容器化部署）

### 安装

#### 方法一：直接安装

1. 克隆仓库

```bash
git clone https://github.com/Yuri-NagaSaki/ImageFlow.git
cd ImageFlow
```

2. 构建前端

```bash
cd frontend
bash build.sh
cd ..
```

3. 构建后端

```bash
go mod tidy
go build -o imageflow
```

4. 配置环境变量

```bash
cp .env.example .env
# 编辑 .env 文件进行配置
```

5. 设置系统服务（以 systemd 为例）

```ini
[Unit]
Description=ImageFlow Service
After=network.target

[Service]
ExecStart=/path/to/imageflow
WorkingDirectory=/path/to/imageflow/directory
Restart=always
User=youruser
EnvironmentFile=/path/to/imageflow/.env

[Install]
WantedBy=multi-user.target
```

6. 启用服务

```bash
sudo systemctl enable imageflow
sudo systemctl start imageflow
```

#### 方法二：Docker 部署

1. 使用预构建镜像（推荐）

```bash
# 1. 克隆仓库
git clone https://github.com/Yuri-NagaSaki/ImageFlow.git
cd ImageFlow

# 2. 配置环境变量
cp .env.example .env
# 编辑 .env 文件

# 3. 启动服务
docker compose up -d
```

2. 本地构建部署

```bash
# 1. 克隆仓库
git clone https://github.com/Yuri-NagaSaki/ImageFlow.git
cd ImageFlow

# 2. 配置环境变量
cp .env.example .env
# 编辑 .env 文件

# 3. 构建并启动
docker compose -f docker-compose-build.yml up --build -d
```

### 配置说明

通过创建和编辑 `.env` 文件配置系统。以下是主要配置项说明：

```bash
# API 密钥配置
API_KEY=your_api_key_here  # 设置您的 API 密钥

# 存储配置
STORAGE_TYPE=local  # 存储类型：local（本地存储）或 s3（S3 兼容存储）
LOCAL_STORAGE_PATH=static/images  # 本地存储路径

# S3 存储配置（当 STORAGE_TYPE=s3 时需要）
S3_ENDPOINT=  # S3 端点地址
S3_REGION=    # S3 区域
S3_ACCESS_KEY=  # 访问密钥
S3_SECRET_KEY=  # 访问密钥密文
S3_BUCKET=      # 存储桶名称
CUSTOM_DOMAIN=  # 自定义域名

# 图像处理配置
MAX_UPLOAD_COUNT=20    # 单次最大上传数量
IMAGE_QUALITY=80      # 图像质量（1-100）
WORKER_THREADS=4      # 并行处理线程数
COMPRESSION_EFFORT=6  # 压缩级别（1-10）
FORCE_LOSSLESS=false  # 是否强制无损压缩
```


## 📝 使用方法

### API 密钥认证

图片上传功能需要 API 密钥认证。您可以：

1. 在 `.env` 文件中设置 API 密钥
2. 通过网页界面输入 API 密钥
3. API 密钥在验证成功后将被保存

### 上传图片

访问 `http://localhost:8686/` 的上传界面。您可以：

1. 将图片拖拽到上传区域
2. 点击选择要上传的图片
3. 选择图片的过期时间（可选）
4. 添加标签对图片进行分类（可选）
4. 实时预览选中的图片
5. 系统自动检测图片是横向还是纵向
6. 上传后，图片会自动转换为 WebP 和 AVIF 格式
7. 如果设置了过期时间，图片将在过期后自动删除

### 管理图片

访问 `http://localhost:8686/manage.html` 的管理界面。您可以：

1. 查看所有已上传的图片，并可按格式、方向和标签进行筛选
2. 点击任意图片查看详细信息
3. 复制图片的直接URL以方便分享
4. 删除不再需要的图片（需要API密钥认证）
5. 当删除图片时，所有相关格式（原始、WebP、AVIF）将同时被移除

### 获取随机图片

通过 API 获取随机图片（无需 API 密钥）：

```
GET http://localhost:8686/api/random
GET http://localhost:8686/api/random?tag=nature
```

系统会根据请求头中的设备类型和浏览器支持返回最合适的图片。您还可以通过标签筛选随机图片。

### API 参考

| 接口 | 方法 | 描述 | 参数 | 认证 |
|----------|---------|-------------|------------|-------------|
| `/auto/image/<original path>` | GET | 自适应获取 Webp 或 AVIF 图片 | 无 | 不需要 |
| `/api/random` | GET | 获取随机图片 | `tag`：可选，按标签筛选<br> | 不需要 |
| `/api/upload` | POST | 上传新图片 | Form 数据，字段名 "images[]"<br>可选参数：`expiryMinutes`（过期时间，分钟）<br>可选参数：`tags`（标签数组） | 需要 API 密钥 |
| `/api/delete-image` | POST | 删除图片及其所有格式 | JSON 数据，包含 `id` 和 `storageType` | 需要 API 密钥 |
| `/api/validate-api-key` | POST | 验证 API 密钥 | 请求头中的 API 密钥 | 不需要 |
| `/api/images` | GET | 列出所有已上传的图片 | 可选：`tag`（按标签筛选） | 需要 API 密钥 |
| `/api/config` | GET | 获取系统配置 | 无 | 需要 API 密钥 |
| `/api/trigger-cleanup` | POST | 手动触发清理过期图片 | 无 | 需要 API 密钥 |
| `/api/tags` | GET | 获取所有可用标签 | 无 | 需要 API 密钥 |

### 项目结构

```
ImageFlow/
├── .github/        # GitHub 相关配置
├── config/         # 配置相关代码
├── docs/          # 文档和图片
├── favicon/       # 网站图标资源
├── frontend/      # Next.js 前端应用
│   ├── app/       # Next.js 应用目录
│   ├── public/    # 公共资源
│   ├── .next/     # Next.js 构建输出
│   ├── out/       # 静态导出输出
│   ├── build.sh   # Unix 构建脚本
│   └── build.bat  # Windows 构建脚本
├── handlers/      # HTTP 请求处理器
├── scripts/       # 实用脚本
├── static/        # 静态文件和图片存储
│   └── images/    # 图片存储目录
│       ├── landscape/  # 横向图片
│       │   ├── avif/   # AVIF 格式
│       │   └── webp/   # WebP 格式
│       ├── portrait/   # 纵向图片
│       │   ├── avif/   # AVIF 格式
│       │   └── webp/   # WebP 格式
│       ├── original/   # 原始图片
│       │   ├── landscape/  # 原始横向图片
│       │   └── portrait/   # 原始纵向图片
│       ├── gif/       # GIF 格式图片
│       └── metadata/  # 图片元数据（包含过期时间信息）
├── utils/         # 实用函数
├── .env          # 环境变量
├── .env.example  # 环境配置示例
├── Dockerfile    # Docker 配置
├── docker-compose.yaml      # Docker Compose 配置
├── docker-compose-build.yml # Docker Compose 构建配置
├── go.mod        # Go 模块文件
├── go.sum        # Go 模块校验和
├── main.go       # 主程序入口
└── README.md     # 项目文档
```

## 🆕 最近更新

### 版本 1.4.1

- **标签管理**：添加了在上传时为图片添加标签和按标签筛选的功能
- **改进的UI**：增强了标签管理界面，采用现代化设计
- **Bug修复**：
  - 修复了图片过期功能，使其在本地和S3存储中都能正常工作
  - 改进了PNG图像处理，支持可选的无损压缩
  - 修复了随机图片API，使其能够正确按标签筛选
  - 优化了设备特定的图像方向检测

## 🤝 贡献

欢迎贡献！随时提交代码、报告问题或提出改进建议！

## 📄 许可证

本项目基于 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。

## 📞 联系方式

Blog - [猫猫博客](https://catcat.blog)

项目链接：[https://github.com/Yuri-NagaSaki/ImageFlow](https://github.com/Yuri-NagaSaki/ImageFlow)

---

## ❤️ Thanks
[YXVM](https://support.nodeget.com/page/promotion?id=80)赞助了本项目

[NodeSupport](https://github.com/NodeSeekDev/NodeSupport)赞助了本项目

<div align="center">
  <p>⭐ 如果您喜欢这个项目，请给它一个星标！⭐</p>
  <p>由 Yuri NagaSaki 用 ❤️ 制作</p>
</div>

#### 部署说明

- 服务默认运行在 8686 端口
- Docker 部署会自动包含所有必要的依赖
- 支持多架构部署（amd64/arm64）
- 提供健康检查功能
- 支持环境变量配置和覆盖

#### 常见问题排查

1. 容器显示 unhealthy：
   - 检查健康检查日志：`docker logs <container_id>`
   - 确保容器内部的 8686 端口正常监听：`docker exec <container_id> netstat -tulpn`
   - 验证服务是否正在运行：`docker exec <container_id> ps aux`

2. 外网无法访问：
   - 确保 SERVER_ADDR 设置为 0.0.0.0:8686
   - 检查防火墙设置：`sudo ufw status` 或 `sudo iptables -L`
   - 验证端口映射：`docker port <container_id>`
   - 尝试使用 host 网络模式：
     ```yaml
     # 在 docker-compose.yaml 中添加
     network_mode: "host"
     ```

3. 静态资源访问问题：
   - 确保卷挂载正确：`docker inspect <container_id>`
   - 检查目录权限：`ls -la ./static/images`
   - 验证静态文件是否存在：`docker exec <container_id> ls -la /app/static`

4. 性能优化建议：
   - 使用 volume 而不是 bind mount 来提高性能
   - 配置适当的 WORKER_THREADS 数量
   - 根据服务器配置调整 IMAGE_QUALITY 和 COMPRESSION_EFFORT
