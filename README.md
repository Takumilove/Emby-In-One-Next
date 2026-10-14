# Emby-In-One-Next

<div align="center">

### 下一代高性能 Emby 多服聚合与播放增强网关
**Unify Multiple Emby Servers into One Seamless, High-Performance Media Gateway**

[![Release](https://img.shields.io/github/v/release/Takumilove/Emby-In-One-Next?color=brightgreen&label=Release&logo=github)](https://github.com/Takumilove/Emby-In-One-Next/releases)
[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker&logoColor=white)](https://www.docker.com/)
[![License: GPL v3](https://img.shields.io/badge/License-GPL%20v3-blue.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20amd64%20%7C%20arm64%20%7C%20armv7-lightgrey)]()

[📖 快速安装](#-快速安装) • [🌟 为什么选择 Next](#-为什么选择-emby-in-one-next) • [🤝 致谢与开源传承](#-致谢与开源传承) • [⚡ 核心特性](#-核心特性概览) • [🛠️ 运维与管理](#-运维与管理) • [English README](README_EN.md)

</div>

---

## 🌟 为什么选择 Emby-In-One-Next？

**Emby-In-One-Next** 是专为影视发烧友打造的多 Emby 服务器聚合与播放增强网关。它可以将你拥有的多个 Emby 影视服（自建服、公益服、网盘挂载服等）聚合为一个统一的入口，让任何官方与第三方 Emby 客户端以单一服务器的形式无缝访问。

#### 相比原版与传统方案，Next 彻底攻克了以下关键痛点：

| 痛点场景 | 传统 / 原版方案表现 | 🚀 Emby-In-One-Next 解决方案 |
| :--- | :--- | :--- |
| **首页「继续观看」** | 批量请求 `/Items` 报 `403 Forbidden`，首页继续观看卡片经常一片空白 | **自适应用户作用域穿透**：优先带 `UserId` 递归查询，失败时智能双向回退；单项逐个并发拉取兜底，保障卡片 100% 呈现 |
| **播放进度条丢失** | 上游缺少 `RunTimeTicks` 时，客户端直接丢弃继续观看卡片或不显进度 | **实时指标动态丰满**：自动补齐时长、海报标签、剧名，并根据本地进度计算 `PlayedPercentage`，完美渲染进度条 |
| **多客户端兼容性** | AfuseKt、Yamby、Infuse、SenPlayer 遇到小写路由或缺失端点频繁报错/断播 | **全路由矩阵兼容**：补齐 `/Users/{userId}/...` 系列与通用端点双向路由；路径大小写容错映射；支持 JSON、Form、Query 多格式鉴权透传 |
| **剧集「接下来观看」** | 无法跨服汇聚剧集未播下一集，连续追剧体验割裂 | **跨服 NextUp 智能汇聚**：自动聚合多服务器未播单集，智能过滤完结/无效条目，无缝顺播 |
| **片源合并与多服选源** | 第三方播放器请求 PlaybackInfo 无法触发多版本合并与并发限制 | **PlaybackInfo 用户端点补齐**：完整支持多服务器片源版本（MediaSources）合并与 `maxConcurrent` 独立并发流控 |
| **网盘服播放开销** | 代理模式占用 VPS 巨额流量，甚至因上游 302 重定向破图/断流 | **网盘 302 直链透传 + 图片跟随**：直链重定向给客户端直连，节省代理服务器宽带；自动追踪图片跳转，杜绝破图 |
| **多用户数据混乱** | 分发给朋友/家人后，观看历史和播放进度被上游公用账号覆盖篡改 | **真正的多用户隔离**：本地 SQLite 数据库独立记录每位普通用户的观看历史、收藏与已播放状态，互不干扰 |
| **部署与历史迁移** | 升级容易造成配置冲突或数据覆盖 | **无感平滑迁移**：内置自动检测并安全迁移旧版 `/opt/emby-in-one` 配置与数据；保留 `emby-in-one` 命令软链接兼容 |

---

## 🤝 致谢与开源传承

**Emby-In-One-Next** 秉承开源精神，深怀感恩之心：

- 本项目基于原项目 [ArizeSky/Emby-In-One](https://github.com/ArizeSky/Emby-In-One)（由原作者 [@chen3861229](https://github.com/chen3861229) 构想与开发）的优秀架构深度演化而来。感谢原作者为社区带来的开创性灵感与最初的 Go 虚拟化底座。
- 感谢社区优秀贡献者 [@ssabv](https://github.com/ssabv) 在网盘服 302 直链透传、分页拉取与图片重定向领域的宝贵代码贡献与探索。
- **Next 版本的使命**：原版项目在解决基础聚合后较少合并社区关于复杂多端播放与深层路由缺陷的反馈。`Emby-In-One-Next` 旨在接棒开源火炬，全面重构修复多客户端断播/漏播问题、深化全功能路由覆盖、确立独立的 **V1.0.0** 独立演进版本线，为所有媒体聚合爱好者提供一个稳定、积极演进、功能完备的次世代解决方案。
- 本项目严格遵循 **GNU General Public License v3.0 (GPL-3.0)** 协议持续开源。

---

## ⚡ 核心特性概览

### 🎬 1. 跨服播放与进度无缝聚合
- **聚合首页「继续观看」（Resume）**：多服务器观看进度统一汇总至首页，进度条精准对齐，点开即播。
- **跨服「接下来观看」（NextUp）**：追剧无需切换服务器，系统自动根据当前追剧进度跨服检索下一集。
- **多版本片源智能合并**：同一电影/剧集在不同服务器有 4K HDR、1080P、蓝光原盘等不同版本时，自动合并至同一影片详情页，供播放器自由选择画质版本。
- **4 级元数据智能择优**：指定标记优先 > 中文优先 > 简介长度 > 优先顺序，自动提取最完整、高质量的海报与刮削简介。

### 📱 2. 全客户端深度兼容与协议适配
- **无缝支持全生态客户端**：Emby 官方各端（Web / Android TV / iOS / macOS / Windows）、AfuseKt、Yamby、Infuse、VidHub、SenPlayer 等。
- **协议与大小写容错**：自动对 Emby 静态路由进行规范化，兼容客户端大小写差异（如 `/Users/authenticatebyname`）；兼容 `Token=` 头格式与 URL Query 传参。
- **单项与推荐路由覆盖**：补齐 `Similar`（相似推荐）、`ThemeMedia`（主题背景音乐/视频）、`Search/Hints`（全局搜索联想词）等。

### 👥 3. 真正的多用户与权限隔离
- **独立普通用户体系**：无需在上游 Emby 开设新账号，网关层即可创建多个普通用户。
- **首页媒体库按需隐藏**：管理员可针对不同用户（包括管理员自己）自定义隐藏指定服务器的特定媒体库，界面清爽整洁。
- **播放历史本地隔离**：普通用户的播放进度、已播放打勾、个人收藏均在本地数据库持久化，彻底与上游公用账号隔离。
- **精细化角色控制**：普通用户受分配服务器限制，严禁访问任何管理接口与非授权上游节点。

### 🚀 4. 高性能、流量节省与自适应流控
- **并发请求与宽恕期机制**：向所有上游并发拉取，快服优先秒开，慢服在可配置宽恕期窗口内汇入，彻底告别“短板效应”。
- **网盘服 302 直链透传（`directRedirect`）**：针对网盘或 STRM 资源，自动解析最终 CDN 播放链接并以 302 重定向直连客户端，VPS 几乎 0 流量消耗。
- **上游分页防爆扫描（`pagedScan`）**：针对防刷限制单次请求条数的魔改服，采用 200 条/批次自动分页扫描，杜绝 400 报错。
- **单服独立最大并发限制（`maxConcurrent`）**：严格控制每台上游的并发播放会话数，超出时友好响应 429，防止上游账号因多播被封禁。
- **多推流备用线路**：每台上游均可配置多条推流备选地址（内网穿透、海外加速等），故障自动降级。

### 🛡️ 5. 安全可靠与高可用
- **非 Root 降权运行**：默认以专属非特权用户（UID 1000）运行，防止容器越权。
- **强密码加密哈希**：管理员密码采用加盐 scrypt 哈希存储，拒绝任何明文泄露。
- **SSRF 深度防御与防爆破限制**：管理面板严格拦截回环与局域网内网探测；登录接口配备连续失败 IP 临时封禁策略。
- **单点故障自动剔除与回退**：若某一台上游失联或故障，网关自动标记并剔除，已聚合项目自动通过备用副本回退，避免界面加载卡死。

### 💻 6. 极简运维与工具生态
- **双端自托管管理面板**：内嵌现代化 Web 管理面板（Vue 3 + Tailwind CSS），全静态编译入二进制，0 外部 CDN 依赖，安全杜绝白屏。
- **交互式 SSH 终端 CLI**：输入 `emby-in-one-next`（或 `emby-in-one`）即可呼出全彩命令行管理菜单，启动/重启/日志/查密/改密/更新一键搞定。
- **单静态可执行文件**：Go 纯静态编译，无 glibc 动态库依赖，全平台开箱即用。

---

## 🖼️ 预览

![预览图1](https://cdn.nodeimage.com/i/D293pIQcFNx4gXkfskPbnXFzmgCQ1JPx.webp)
![预览图2](https://cdn.nodeimage.com/i/iDAXrYaIXdm9efhwl2BtqJjRUmGfTSKU.webp)
![预览图3](https://cdn.nodeimage.com/i/K4jhTTMjv8rkHYiPNbXKUC0kXIzAXgq0.webp)
![预览图4](https://cdn.nodeimage.com/i/jCilzHTw7vzRJYaQFtbvd8ZOEaTxZvk6.webp)

---

## 📖 快速安装

本项目推荐使用 **Release 二进制一键脚本** 部署（无需本地编译 Go，直接下载对应架构的高性能纯静态二进制，开箱即用）；也可选择 **Docker Compose** 容器化部署。

### 方式一：Release 二进制一键安装（强烈推荐）

适用于各种主流 Linux 发行版（Debian、Ubuntu、CentOS、AlmaLinux 等）：

```bash
curl -fsSL -o release-install.sh https://raw.githubusercontent.com/Takumilove/Emby-In-One-Next/main/release-install.sh
sudo bash release-install.sh
```

该脚本全自动执行：
- 识别当前系统的 CPU 架构（amd64 / arm64 / armv7 / mips / mipsle / riscv64），自动拉取最新 Release 纯静态二进制；
- 建立标准部署目录 `/opt/emby-in-one-next`，若检测到旧版 `/opt/emby-in-one`，**自动平滑迁移已有数据与配置**；
- 生成随机管理员密码并采用 scrypt 加盐加密存储；
- 注册 `emby-in-one-next.service` 系统服务并设置开机自启；
- 安装 `emby-in-one-next` 管理命令（并建立 `emby-in-one` 兼容软链接）。

### 方式二：Docker Compose 容器化部署

1. 创建项目目录并授予对应权限（容器以 uid 1000 运行）：
```bash
mkdir -p /opt/emby-in-one-next/{config,data}
chown -R 1000:1000 /opt/emby-in-one-next/config /opt/emby-in-one-next/data
cd /opt/emby-in-one-next
```

2. 编写 `docker-compose.yml`：
```yaml
services:
  emby-in-one-next:
    image: ghcr.io/takumilove/emby-in-one-next:latest
    container_name: emby-in-one-next
    user: "1000:1000"
    ports:
      - "8096:8096"
    volumes:
      - ./config:/app/config
      - ./data:/app/data
    restart: unless-stopped
    security_opt:
      - no-new-privileges:true
    logging:
      driver: json-file
      options:
        max-size: "10m"
        max-file: "3"
```

3. 启动容器：
```bash
docker compose up -d
```

### 方式三：Go 源码编译运行（开发者模式）

需要本地具备 Go 1.23+ 环境及 C 编译工具链（用于编译 SQLite CGO）：

```bash
git clone https://github.com/Takumilove/Emby-In-One-Next.git
cd Emby-In-One-Next
mkdir -p config data
go run ./cmd/emby-in-one-next
```

### 访问地址与初始登录

- **Emby 客户端连接地址**：`http://<服务器IP>:8096`
- **Web 管理后台面板**：`http://<服务器IP>:8096/admin`
- **默认用户名**：`admin`
- **初始随机密码**：一键安装完成后会在终端打印；若是手动部署未预设密码，首次启动会自动生成。如需重置，运行交互式 SSH 菜单即可。

---

## 🛠️ 运维与管理

### 交互式 SSH 管理菜单

在终端直接运行：
```bash
emby-in-one-next
# 或者输入原有简写命令
emby-in-one
```

菜单功能包括：
1. **服务启停控制**：启动、停止、重启服务（自动适配 systemd 或 docker compose）
2. **在线升级**：一键检查并更新到最新 Release 版本（保留所有配置与历史数据）
3. **账号安全维护**：查看管理员账号、重置管理员密码、修改管理员用户名
4. **多用户管理**：快速查看用户列表、添加/删除普通分发用户
5. **实时日志排障**：实时滚动追踪最近 50 条关键日志
6. **一键干净卸载**：支持选择保留或彻底清理数据目录

---

## ⚙️ 配置文件说明

配置文件位于 `config/config.yaml`（在 Web 管理面板修改大多数选项可**热生效**，无需重启服务）。

```yaml
# dataDir: "/opt/emby-in-one-next/data"    # 运行时数据目录（顶层键，默认值见下方「数据目录」说明）

server:
  port: 8096
  name: "Emby-In-One-Next"
  # id: 首次启动自动生成，请勿手动修改
  # trustProxy: true        # 部署在反向代理（Nginx/Caddy/Cloudflare 等）后面时务必设为 true

admin:
  username: "admin"
  password: "your-strong-password"    # 首次启动后自动加密为 scrypt 哈希存储

playback:
  mode: "proxy"          # "proxy" 或 "redirect"，全局默认值

timeouts:
  api: 30000             # 单次上游 API 请求超时（ms）
  global: 15000          # 聚合请求总超时——等待所有服务器的最大时长（ms）
  login: 30000           # 上游登录超时（ms）
  healthCheck: 30000     # 健康检查超时（ms）
  healthInterval: 60000  # 健康检查间隔（ms）
  searchGracePeriod: 3000     # 搜索聚合宽恕期（ms），0 表示禁用
  metadataGracePeriod: 3000   # 元数据获取宽恕期（ms），0 表示禁用
  latestGracePeriod: 0        # "最新添加"宽恕期——0 表示等待全部服务器（ms）

proxies: []
  # - id: "proxy-hk"
  #   name: "香港节点"
  #   url: "http://user:pass@ip:port"

upstream:
  - name: "主站服务器"
    url: "https://emby.example.com"
    username: "user"
    password: "pass"
    # 或者使用 API Key:
    # apiKey: "your-api-key"
    playbackMode: "proxy"                      # proxy | redirect
    spoofClient: "none"                        # none | passthrough | infuse | custom
    streamingUrls:                             # 推流线路（支持主线路+备用线路）
      - "https://cdn.example.com"
      - "https://backup.example.com"
    followRedirects: true                      # 是否跟随 302 重定向
    directRedirect: true                       # 开启网盘/STRM 媒体直链 302 透传（节约带宽）
    pagedScan: false                           # 开启分页分批拉取（单批 200 条，防 400）
    maxConcurrent: 3                           # 最大并发播放数，0 为不限
    priorityMetadata: false                    # 发生去重冲突时，优先采用此服的元数据与海报
```

---

## 👥 多用户管理与独立观看数据

### 角色权限

| 角色 | 权限与范围 | 观看历史与进度 |
| :--- | :--- | :--- |
| **管理员 (admin)** | 访问所有媒体库、访问管理面板、执行管理 API | 与上游公用账号保持一致 |
| **普通用户 (user)** | 仅允许访问管理员分配的指定服务器；禁止访问管理后台 | **本地独立数据库双写记录**，隔离在上游账号之外 |

### 独立观看数据工作机制

1. **继续观看与接下来播放**：用户在任何播放器上的进度上报，网关均会在本地 SQLite (`mappings.db`) 实时双写持久化，首页根据该用户的本地进度精准计算。
2. **收藏与已播放状态**：支持客户端直接标记已看/未看/收藏，仅作用于该用户自身，不会污染其他用户或上游。
3. **首页库按需隐藏**：管理面板可为不同用户单独配置屏蔽某些上游库，满足不同家庭成员的观看偏好。

---

## 🛡️ 安全加固与排障

### 反向代理信任 (`trustProxy`)
如果网关部署在反向代理（Nginx / Caddy / Cloudflare 等）之后，**务必在配置中开启 `trustProxy: true`**。并且请确保反向代理**覆写**（Overwrite）而非追加 `X-Real-IP` 或 `X-Forwarded-For` 请求头：
```nginx
proxy_set_header X-Real-IP $remote_addr;
proxy_set_header X-Forwarded-For $remote_addr;
```
这样能防止客户端通过伪造请求头恶意绕过 15 分钟登录失败防爆破限制。

### 日志系统
- **日志文件**：默认存储于 `data/emby-in-one-next.log`（单文件 10MB，自动保留 3 份轮转备份）。
- **查看日志**：可通过 SSH 菜单选项 `10` 实时查看，或在 Web 管理面板的「系统日志」页一键导出/清空。

---

## 🏗️ 项目架构

```text
Emby-In-One-Next/
├── cmd/emby-in-one-next/           # 程序主入口与命令行工具
├── internal/backend/               # 核心网关与代理逻辑
│   ├── routes.go                   # 完整路由注册与规范化拦截
│   ├── media.go                    # 媒体汇聚、智能去重与元数据优先级处理
│   ├── media_resume.go             # 跨服「继续观看」智能拉取与单项逐个兜底
│   ├── media_nextup.go             # 跨服「接下来观看」单集计算与双路由
│   ├── media_playback.go           # PlaybackInfo 代理与多版本片源合并、并发流控
│   ├── streamproxy.go              # 流媒体高可用代理、网盘 302 透传与 HLS 重写
│   ├── idstore.go                  # SQLite 双向虚拟 ID 映射持久化
│   ├── watch_store.go              # 多用户独立观看进度数据库引擎
│   ├── identity.go                 # 客户端身份识别与 Passthrough 五级伪装
│   ├── ssrf.go                     # SSRF 安全防御与 DNS 重新绑定防护
│   └── logger.go                   # 结构化多级别双输出滚动日志引擎
├── public/                         # 自托管内嵌 Web 管理面板（Vue 3 + Tailwind CSS）
├── Dockerfile                      # 多阶段静态编译镜像构建文件
├── docker-compose.yml              # 容器编排定义
├── install.sh                      # 源码构建一键安装脚本
├── release-install.sh              # 预编译静态二进制一键部署脚本
└── emby-in-one-next-cli.sh         # 交互式 SSH 终端管理脚本
```

---

## 📄 免责声明

本项目仅供个人学习、研究及多媒体库技术交流使用。请务必遵守上游服务器的服务条款与所在地区的法律法规。由于模拟客户端行为导致的上游封禁或账号风险，由使用者自行承担。

---

## 📜 许可证

本项目采用 [GNU General Public License v3.0 (GPL-3.0)](LICENSE) 协议开源。
