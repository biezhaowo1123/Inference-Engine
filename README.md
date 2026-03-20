# Inference Engine

基于大模型 API 的智能进程发展推理平台，内置 24 个模型配置入口，提供多步推理、多方案对比、可视化展示等功能。

[English](#english) | [中文](#中文)

---

## 中文

### 功能特点

- **多步推理** - 分步骤推演进程发展，每步包含标题、描述、推理逻辑和置信度
- **多模型支持** - 支持按配置启用 DeepSeek、Claude、GPT、Gemini、通义千问、智谱GLM、MiniMax 等 24 个模型入口
- **多方案对比** - 生成多种可能的发展路径
- **可视化展示** - 清晰的时间线图表展示推理过程
- **历史记录** - 保存推理历史，支持删除和清空

### 支持的模型

#### 国际主流模型

| 模型 | 提供商 | 特点 | 价格参考 |
|------|--------|------|----------|
| DeepSeek | 深度求索 | 中文能力强，性价比极高 | ¥1/百万token |
| Claude | Anthropic | 推理能力强，适合复杂分析 | $3/百万token |
| GPT-4o | OpenAI | 综合能力强，生态完善 | $5/百万token |
| Gemini | Google | 免费，多模态能力强 | 免费 |
| Grok | xAI | 实时性强 | 早期免费 |
| Mistral | Mistral AI | 欧洲开源模型，性能优秀 | $2/百万token |

#### 国内大厂模型

| 模型 | 提供商 | 特点 | 价格参考 |
|------|--------|------|----------|
| 通义千问 | 阿里云 | 中文能力强，国内稳定 | ¥0.8/千token |
| 智谱GLM | 智谱AI | 清华背景，中文能力强 | ¥0.1/千token |
| 文心一言 | 百度 | 百度大模型 | ¥0.12/千token |
| 混元 | 腾讯 | 多模态能力 | Lite版免费 |
| 星火 | 讯飞 | 语音+文本能力 | ¥0.036/千token |
| 豆包 | 字节跳动 | 极便宜 | ¥0.0008/千token |

#### 国内创业公司模型

| 模型 | 提供商 | 特点 | 价格参考 |
|------|--------|------|----------|
| MiniMax | MiniMax | 语音合成能力强 | ¥0.03/千token |
| Moonshot | 月之暗面 | 长文本能力强 | ¥0.012/千token |
| 百川 | 百川智能 | 中文能力强 | ¥0.012/千token |
| 零一万物 | 李开复团队 | 性价比高 | ¥0.006/千token |

#### 本地模型（完全免费）

| 模型 | 特点 |
|------|------|
| Ollama | 完全免费，隐私安全，无限制 |
| vLLM | 高性能推理，支持多种开源模型 |
| LocalAI | OpenAI 兼容的本地推理 |

### 快速开始

#### 1. 克隆项目

```bash
git clone https://github.com/biezhaowo1123/Inference-Engine.git
cd Inference-Engine
```

#### 2. 配置 API Key

```bash
# 复制配置文件
cp configs/.env.example .env

# 编辑配置文件，填入你的 API Key
# 至少启用一个模型，并填入对应 API Key
# 同时把对应的 *_ENABLED 改成 true
vim .env
```

配置说明:

- `DEFAULT_MODEL` 是请求体未显式传入 `model` 时使用的默认模型
- 创建推理页面只会展示当前已启用的模型
- 需要 API Key 的模型，除了 `*_ENABLED=true` 之外，还需要配置对应密钥
- `ollama`、`vllm`、`localai` 这类本地 OpenAI 兼容接口可以不填 API Key

#### 3. 运行

```bash
# 安装依赖
go mod tidy

# 启动服务
go run ./cmd/server
```

#### 4. 访问

打开浏览器访问 http://localhost:8080

### 使用方法

1. 访问首页，点击「开始推理」
2. 选择推理领域（历史/商业/技术/个人）
3. 输入推理主体和关键变化点
4. 设置时间范围和推理步数
5. 选择 AI 模型
6. 点击「开始推理」

说明:

- 如果请求中不传 `model`，后端会回退到 `.env` 里的 `DEFAULT_MODEL`
- 页面中的模型列表由后端动态返回，只显示当前真实可用的模型

### 示例场景

#### 历史推演
- 如果扶苏继位，秦朝会怎样发展？
- 如果明朝没有海禁，中国历史会如何改变？

#### 商业预测
- 如果公司坚持自主研发，未来5年会怎样？
- 如果产品定价策略调整，市场份额会如何变化？

#### 技术发展
- 如果 AI 持续发展，2030年编程会变成什么样？
- 如果新能源技术突破，能源格局会如何变化？

#### 个人决策
- 如果我选择创业而不是打工，3年后会怎样？
- 如果我学习新技术，职业发展会如何？

### 技术架构

```
inference-engine/
├── cmd/server/          # 服务入口
├── configs/             # 配置文件示例
├── internal/
│   ├── api/             # API 层 (Gin)
│   ├── config/          # 配置管理
│   ├── engine/          # 推理引擎核心
│   ├── models/          # 数据模型
│   └── storage/         # 数据存储 (SQLite/PostgreSQL)
├── web/
│   ├── static/          # 静态资源 (CSS)
│   └── templates/       # HTML 模板
└── README.md
```

### API 接口

#### 创建推理任务

```http
POST /api/inference
Content-Type: application/json

{
    "title": "推理标题",
    "domain": "历史",
    "subject": "秦朝",
    "change_point": "扶苏继位",
    "time_frame": {"start": "前210年", "end": "前180年"},
    "steps_count": 5
}
```

`model` 字段可选；如果省略，则使用 `DEFAULT_MODEL`。

当前代码状态补充说明:

- OpenAI-compatible 路线已经接通，可用于 DeepSeek、GPT、Qwen、MiniMax、OpenRouter、Ollama、vLLM、LocalAI 等接口风格兼容的提供商
- Claude、Gemini、文心一言、Cohere 这些专用适配器目前还是占位实现，适合继续开发后再用于真实生产调用

#### 获取推理结果

```http
GET /api/inference/:id
```

#### 获取历史记录

```http
GET /api/history
```

#### 删除历史记录

```http
DELETE /api/history/:id
```

#### 清空历史记录

```http
DELETE /api/history
```

#### 获取可用模型

```http
GET /api/models
```

### 开发计划

- [ ] 流式输出支持
- [ ] 多方案并发对比
- [ ] 可视化时间线图表
- [ ] 用户认证系统
- [ ] API 开放平台
- [ ] 移动端适配
- [ ] Docker 部署支持

### 技术栈

- **后端**: Go 1.21+, Gin, GORM
- **前端**: HTML/CSS/JavaScript (原生)
- **数据库**: SQLite (默认) / PostgreSQL

### License

MIT License

---

## English

An intelligent process inference platform based on LLM APIs with 24 built-in model configuration entries, supporting multi-step reasoning, multi-scenario comparison, and visualization.

### Features

- **Multi-step Reasoning** - Step-by-step process inference with title, description, reasoning logic, and confidence
- **Multi-model Support** - Support for enabling DeepSeek, Claude, GPT, Gemini, Qwen, GLM, MiniMax, and 24 model entries through configuration
- **Multi-scenario Comparison** - Generate multiple possible development paths
- **Visualization** - Clear timeline display of inference process
- **History Management** - Save inference history with delete and clear functions

### Quick Start

```bash
# Clone the repository
git clone https://github.com/biezhaowo1123/Inference-Engine.git
cd Inference-Engine

# Configure API Key
cp configs/.env.example .env
# Edit .env, enable at least one model, and add its API key if required

# Run
go mod tidy
go run ./cmd/server

# Visit http://localhost:8080
```

### API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | /api/inference | Create inference task |
| GET | /api/inference/:id | Get inference result |
| GET | /api/history | Get history list |
| DELETE | /api/history/:id | Delete single history |
| DELETE | /api/history | Clear all history |
| GET | /api/models | Get available models |

Notes:

- If `model` is omitted in `POST /api/inference`, the backend falls back to `DEFAULT_MODEL`
- The inference page only renders models that are actually enabled and available
- OpenAI-compatible providers are wired up; some provider-specific adapters are still placeholders and need additional implementation

### Tech Stack

- **Backend**: Go 1.21+, Gin, GORM
- **Frontend**: HTML/CSS/JavaScript (Vanilla)
- **Database**: SQLite (default) / PostgreSQL

### License

MIT License
