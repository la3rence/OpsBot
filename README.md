# OpsBot 🤖️

Serverless + GitHub API 实现基于 Issue 的 Ops 机器人.

类似于 Kubernetes Prow 机器人的白嫖实现。

目前初始化阶段。开发路线图：

1. 定制标签触发相应 Labels 的自动添加或移除 (已实现)
2. 自动评论回复一些特定内容 (基础实现)
3. 基于回复自动关闭一些 Issue 或 PR
4. 结合第三方平台实现 CI
5. 支持可配置多仓库使用 (可直接配置)

## 配置方式

GitHub 仓库 Settings -> WebHook: 新增一个 WebHook，勾选 application/json, all event.

[使用 Vercel 部署](https://go.lawrenceli.me/deploy-opsbot)

Payload URL (即 WebHook Serverless Function API) 为：`https://xxxx.vercel.app/api/index`

需要注册一个新 GitHub 账号并[创建 Personal Access Token](https://github.com/settings/tokens/new)

然后以 `BOT_TOKEN` 作为将上述 Token 环境变量配置到生产环境.

更多参考请[联系作者](https://go.lawrenceli.me/contact)

@MIT 2021 Lawrence