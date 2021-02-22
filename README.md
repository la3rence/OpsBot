[![Test](https://github.com/Lonor/OpsBot/actions/workflows/test.yaml/badge.svg)](https://github.com/Lonor/OpsBot/actions/workflows/test.yaml)
[![codecov](https://codecov.io/gh/Lonor/OpsBot/branch/main/graph/badge.svg?token=H16BEN675E)](https://codecov.io/gh/Lonor/OpsBot)

# OpsBot 🤖️

Serverless + GitHub API 实现基于 Issue / PR 的 Ops 机器人.

类似于 Kubernetes Prow 机器人的白嫖实现。

目前初始化阶段。开发路线图：

1. 定制标签触发相应 Labels 的自动添加或移除 (已实现)
2. 自动评论回复一些特定内容 (基础实现)
3. 基于回复自动关闭/开启 Issue 或 PR
4. 结合第三方平台实现 CI
5. 支持可配置多仓库使用 (可直接配置)

## 已实现的功能

`/label [标签]`       添加一个 label 到某个 issue / PR

`/un-label [移除标签]` 移除某个 issue / PR 的 label

`/close`              关闭 issue / PR

`/reopen`             重新开启 issue / PR

`/approve`            审核通过某个 PR

`/lgtm`               合并某个 PR

## 配置方式

GitHub 仓库 Settings -> WebHook: 新增一个 WebHook，勾选 application/json, all event.

[使用 Vercel 部署](https://go.lawrenceli.me/deploy-opsbot)

Payload URL (即 WebHook Serverless Function API) 为：`https://xxxx.vercel.app/api/index`

需注册一个新 GitHub 账号作为机器人并[创建 Personal Access Token](https://github.com/settings/tokens/new)

然后以 `BOT_TOKEN` 作为将上述 Token 环境变量配置到生产环境。需要邀请此账号作为仓库的 collaborator.

更多参考请[联系作者](https://go.lawrenceli.me/contact)

@MIT 2021 Lawrence