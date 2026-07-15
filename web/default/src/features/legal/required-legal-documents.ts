/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import {
  REQUIRED_LEGAL_LINKS,
  type RequiredLegalDocumentKey,
} from './required-legal-links'

const REQUIRED_LEGAL_CONTENT = {
  'terms-of-service': `**生效日期：2026 年 7 月 15 日**

本条款适用于 Sisyphus 网站、控制台和 API 服务，以中文版本为准。注册、登录、充值或使用服务即表示你同意本条款及同时展示的其他政策。

> **地区限制：中国大陆地区不在本服务支持范围内。本服务不向位于中国大陆地区的个人或组织提供注册、充值、登录或 API 调用服务。**

## 1. 账号

- 你应具备所在地区法律要求的使用资格，并提供真实、有效的注册信息。
- 账号、密码和 API 令牌由你负责保管。因共享、转售或泄露凭证产生的后果由账号持有人承担。
- 未经允许，不得批量注册、转让账号、出租令牌或转售接口。

## 2. 服务与计费

- 可用模型、渠道、价格、倍率和限制以请求发生时的页面及系统记录为准。
- 请求一旦被上游接收，可能因超时、断线或重试产生费用。除重复扣费、经核实的平台故障或法律另有规定外，已消耗额度不予退还。
- 本服务依赖第三方模型和基础设施，不保证任何模型或功能永久可用，也不保证服务始终连续、无错误。

## 3. 内容与风险

你应确保有权提交相关文本、图片、音频、视频和文件。请求内容会被发送给对应模型服务提供方。模型输出可能错误或不完整，使用前应自行核验，不得将其作为高风险决定的唯一依据。

## 4. 暂停与终止

对违法使用、地区规避、欠费、攻击、异常流量、凭证泄露或其他违反政策的行为，本服务可以限制调用、禁用令牌、冻结或终止账号。紧急安全情况下可先处置后通知。

## 5. 条款变更

本服务可根据法律、上游规则或运营需要更新条款、价格和功能。更新内容公布后继续使用服务，视为接受更新后的规则。`,
  'usage-policy': `**生效日期：2026 年 7 月 15 日**

你必须依法使用 Sisyphus，并对输入内容、调用方式和输出用途负责。禁止以下行为：

- 使用 VPN、代理、虚假地址、代注册、多账号或其他方式规避中国大陆地区限制及其他访问控制；
- 欺诈、诈骗、洗钱、身份冒用、违法色情、暴力伤害、骚扰或侵犯他人合法权益；
- 制作或传播恶意软件、钓鱼工具，实施入侵、撞库、攻击、漏洞利用或凭证窃取；
- 批量注册、刷量、垃圾消息、高频重试，或规避速率、计费、内容审核和安全机制；
- 未经授权收集、处理或公开个人信息、机密资料、账号凭证和受保护内容；
- 未经允许共享、出租、转售账号、令牌或接口；
- 在医疗、法律、金融、招聘、公共安全等高风险场景中，将模型输出作为未经人工复核的唯一决定依据。

发现上述行为时，本服务可以立即限速、禁用令牌、冻结或终止账号，并在必要时保留记录或依法处理。`,
  'supported-countries': `**生效日期：2026 年 7 月 15 日**

> **Sisyphus 不向中国大陆地区提供服务。位于中国大陆地区的个人或组织不得注册、充值、登录或调用 API，也不得委托他人或通过 VPN、代理、虚假信息等方式规避限制。**

除中国大陆地区外，服务仅在同时满足以下条件的国家和地区提供：

- 当地法律允许使用相关 AI 和互联网服务；
- 本服务及所选模型提供方支持该地区；
- 用户不受适用的制裁、出口管制或其他法律限制。

模型、支付方式和功能可能因地区或上游政策不同而变化。注册或充值前，用户应自行确认所在地及使用行为符合当地法律和第三方模型规则。本服务可因法律、安全或上游变化随时调整支持范围。`,
  'service-terms': `**生效日期：2026 年 7 月 15 日**

本文件是《服务条款》的补充，适用于不同模型、渠道和媒体能力。

- **第三方模型**：请求会发送至所选模型服务提供方。不同渠道的版本、内容审核、数据处理、上下文和可用性可能不同，用户应同时遵守对应第三方规则。
- **代码与工具**：代码、联网搜索和工具调用可能出错或触发外部操作。用于生产前必须审核，不得让模型未经确认执行付款、删除数据或修改权限等高风险操作。
- **图片、音频、视频和文件**：用户应拥有输入素材及相关肖像、声音、作品和数据的合法权利，不得提交违法、侵权、私密或未授权内容。
- **流式响应与重试**：断线、超时和自动重试可能产生部分输出、重复调用及额外费用，客户端应自行处理幂等和错误恢复。
- **预览功能**：测试或实验功能可能随时变更、限流或下线，不应在没有替代方案时用于关键业务。

模型能力、价格和计费规则以调用时的系统配置为准。特定功能与本文件冲突时适用本文件，其他事项适用《服务条款》《使用政策》《支持的国家和地区》及《隐私政策》。`,
} satisfies Record<RequiredLegalDocumentKey, string>

export function getRequiredLegalDocument(key: RequiredLegalDocumentKey) {
  const link = REQUIRED_LEGAL_LINKS.find((document) => document.key === key)
  return link ? { ...link, content: REQUIRED_LEGAL_CONTENT[key] } : undefined
}
