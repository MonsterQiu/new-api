# Sisyphus API 使用说明

从创建 API Key 到配置聊天客户端、AI 编程工具和开发接口，本说明覆盖最常用的中转接入流程。

## 一分钟快速开始

1. **注册登录**：打开官网，完成账号注册或登录。
2. **准备余额**：充值或兑换额度，并确认余额可用。
3. **创建密钥**：进入 API Key 页面，创建并立即保存完整密钥。
4. **填写客户端**：填写 Base URL、API Key 和模型名称。
5. **发送测试**：新建会话并发送一条简短消息。

> **最重要的两个参数**
>
> OpenAI 兼容 Base URL：`{{API_BASE_URL}}`
>
> API Key：在控制台的 API Key 页面创建。

## 1. 账号、余额与 API Key

API Key 是客户端访问中转服务的凭证。建议为不同客户端分别创建密钥，便于查看用量、设置额度和单独停用。

### 注册与准备余额

1. 进入官网并完成注册或登录。
2. 在控制台充值或兑换额度。
3. 使用前检查余额是否到账。

### 创建 API Key

1. 打开控制台的 **API Key** 页面。
2. 点击创建密钥。
3. 设置容易识别的名称，例如 `Cherry Studio` 或 `Codex`。
4. 按需要设置模型范围、有效期和额度上限。
5. 保存后立即复制完整密钥。

> **安全提醒**
>
> 密钥相当于账户密码。请勿把完整密钥放进截图、群聊或公开代码仓库；怀疑泄露时应立即删除旧密钥。

<!-- 上传图片后删除本注释的首尾标记：
![在控制台创建 API Key](/documentation/images/01-create-api-key.webp)
-->

## 2. 通用配置参数

不同软件的字段名称可能略有差异，但核心参数始终相同。

| 配置项     | 填写内容                   | 说明                                 |
| ---------- | -------------------------- | ------------------------------------ |
| 供应商类型 | OpenAI / OpenAI Compatible | 大多数聊天客户端选择 OpenAI 兼容类型 |
| 站点根地址 | `{{API_ROOT_URL}}`         | 仅在软件会自动追加 `/v1` 时使用      |
| Base URL   | `{{API_BASE_URL}}`         | 不要重复添加 `/v1`                   |
| API Key    | `sk-你的完整密钥`          | 从控制台创建并复制                   |
| 模型名称   | 以模型广场为准             | 必须与模型 ID 完全一致               |

### Base URL 怎么填

- 字段名称是 **Base URL**、**API Base** 或 **OpenAI Endpoint**：通常填写 `{{API_BASE_URL}}`。
- 字段名称是 **API Host**，并明确说明软件会自动添加 `/v1`：填写 `{{API_ROOT_URL}}`。
- 不要把完整请求路径 `/v1/chat/completions` 填进 Base URL。

## 3. 聊天客户端配置

适用于 Cherry Studio、Chatbox、NextChat、Open WebUI 等支持自定义 OpenAI 接口的客户端。

### 通用步骤

1. 打开客户端设置，找到“模型服务”或“模型供应商”。
2. 添加自定义供应商，类型选择 **OpenAI** 或 **OpenAI Compatible**。
3. Base URL 填写 `{{API_BASE_URL}}`。
4. 粘贴控制台创建的 API Key。
5. 使用“获取模型”功能，或手动填写模型广场中的模型 ID。
6. 保存并启用供应商。
7. 新建对话，发送一条简短消息测试。

### Cherry Studio

1. 进入 **设置 → 模型服务**。
2. 添加 OpenAI 类型的自定义供应商。
3. 填写 API Key 和 `{{API_BASE_URL}}`。
4. 点击获取模型；如果获取失败，可以手动添加模型 ID。
5. 启用需要使用的模型后返回聊天页面测试。

<!-- 上传图片后删除本注释的首尾标记：
![Cherry Studio 配置](/documentation/images/02-chat-client-config.webp)
-->

## 4. Codex 与 Claude Code

推荐从控制台 API Key 页面的操作菜单使用 **导入到 CC Switch**，这样可以减少手动编辑配置文件造成的错误。

### 通过 CC Switch 导入

1. 为编程工具单独创建 API Key，并设置合理额度。
2. 打开密钥操作菜单，选择导入到 CC Switch。
3. 选择 Codex 或 Claude，并选择当前可用的主模型。
4. 在 CC Switch 中确认、保存并启用供应商。
5. 完全退出并重新打开对应客户端。

<!-- 上传图片后删除本注释的首尾标记：
![导入 CC Switch](/documentation/images/03-cc-switch-import.webp)
-->

### Codex CLI 手动配置

Codex 用户级配置位于 `~/.codex/config.toml`。所选模型和渠道必须支持 Responses API。

```toml
model = "your-model-id"
model_provider = "gateway"

[model_providers.gateway]
name = "{{SYSTEM_NAME}}"
base_url = "{{API_BASE_URL}}"
env_key = "GATEWAY_API_KEY"
wire_api = "responses"
```

在终端设置密钥：

```bash
export GATEWAY_API_KEY="sk-你的完整密钥"
codex
```

不要把真实密钥直接写入 `config.toml` 或提交到代码仓库。

## 5. 开发者 API 调用

下面示例中的 `your-model-id` 必须替换为模型广场当前展示的模型 ID。

### cURL

```bash
curl "{{API_BASE_URL}}/chat/completions" \
  -H "Authorization: Bearer sk-你的完整密钥" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "your-model-id",
    "messages": [
      {"role": "user", "content": "你好"}
    ]
  }'
```

### Python

```python
from openai import OpenAI

client = OpenAI(
    api_key="sk-你的完整密钥",
    base_url="{{API_BASE_URL}}",
)

response = client.chat.completions.create(
    model="your-model-id",
    messages=[{"role": "user", "content": "你好"}],
)

print(response.choices[0].message.content)
```

### Node.js

```javascript
import OpenAI from 'openai'

const client = new OpenAI({
  apiKey: 'sk-你的完整密钥',
  baseURL: '{{API_BASE_URL}}',
})

const response = await client.chat.completions.create({
  model: 'your-model-id',
  messages: [{ role: 'user', content: '你好' }],
})

console.log(response.choices[0].message.content)
```

> 生产环境请通过环境变量或密钥管理服务保存 API Key，不要将密钥放在浏览器前端代码中。



## 6. 模型选择与计费

- **模型 ID**：以模型广场展示为准，名称和大小写必须完全一致。
- **价格**：不同模型可能按输入、输出、图片、时长或任务次数计费。
- **分组**：API Key 所属分组会影响可用渠道和倍率。
- **用量日志**：可查看模型、令牌数、耗时和扣费明细。
- **余额不足**：请求可能返回 429 或额度不足提示。

<!-- 上传图片后删除本注释的首尾标记：
![模型与价格](/documentation/images/04-models-and-pricing.webp)
-->

## 7. 常见问题排查

### 401：未授权或 API Key 无效

重新复制完整密钥，检查是否包含空格、换行或缺少 `sk-` 前缀，并确认密钥未被禁用。

### 403：没有权限使用模型

检查 API Key 的模型限制、分组权限和账号状态，并确认模型对当前分组开放。

### 404：接口或模型不存在

检查 Base URL 是否重复添加 `/v1`，并确认模型 ID 与模型广场完全一致。

### 429：额度不足或请求过快

查看余额和密钥额度，降低并发与请求频率；如果上游繁忙，请稍后重试。

### 请求超时或中途断开

检查本地网络、代理和防火墙，适当增加客户端超时时间，并尝试其他可用模型。

### 无法获取模型列表

可以先手动添加模型 ID。部分客户端的模型列表接口或鉴权方式可能与自定义中转不完全兼容。

## 8. API Key 安全建议

- 不要在截图、群聊、工单或公开视频中暴露完整密钥。
- 不要把密钥提交到 GitHub、Gitee 等公开代码仓库。
- 为不同设备和客户端创建独立密钥，并设置额度或模型限制。
- 定期查看使用日志，发现异常调用后立即删除旧密钥。
- 服务端通过环境变量保存密钥，浏览器前端不要持有长期密钥。

## 9. 最终检查清单

- [ ] 账号余额可用
- [ ] API Key 已创建且未禁用
- [ ] Base URL 没有重复的 `/v1`
- [ ] 模型 ID 与模型广场完全一致
- [ ] 客户端供应商已经启用
- [ ] 已发送简短消息完成测试

配置完成后，OpenAI 兼容 Base URL 应为：`{{API_BASE_URL}}`
