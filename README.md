[![Test](https://github.com/Lonor/OpsBot/actions/workflows/test.yaml/badge.svg)](https://github.com/Lonor/OpsBot/actions/workflows/test.yaml)
[![codecov](https://codecov.io/gh/Lonor/OpsBot/branch/main/graph/badge.svg?token=H16BEN675E)](https://codecov.io/gh/Lonor/OpsBot)

# OpsBot 🤖️

A robot based on GitHub sdk
and [Vercel's Serverless Function (Go)](https://vercel.com/docs/runtimes#official-runtimes/go). It acts like
the [Kubernetes Prow Robot](https://github.com/k8s-ci-robot). The robot manages your GitHub repo's issues and pull
requests by the content of comments that the user sends.

This project is just getting start and is a toy tool now. For the effectiveness, you can check out any issue or pr from
this repo. If you're interested in this stuff as well, issues or pull requests are welcomed.

## Roadmap / Usage

- [x] `/label [label-name]`    Add a label to the issue / PR
- [x] `/un-label [label-name]` Remove label from the issue / PR
- [x] `/close`                 Close issue / PR
- [x] `/reopen`                Reopen issue / PR
- [x] `/approve`               Approve the PR
- [x] `/lgtm`                  Merge the PR with rebase
- [ ] `/test`                  Test the PR with continuous integration
- [ ] `/assign [username]`     Assign the issue / PR to the user

## Configuration

[Deploy this on Vercel](https://go.lawrenceli.me/deploy-opsbot)

Go to GitHub Repository -> Settings -> WebHook. Add a new WebHook, check `application/json`
and choose all events(or events you care about). Input Payload URL (WebHook Serverless Function API), such
as `https://your-username.vercel.app/api/index`. Protect this webhook by using secret with any string.

Register a new GitHub account (as the robot)
and [create its personal access token](https://github.com/settings/tokens/new). Don't forget to config the `BOT_TOKEN`
and the `WEBHOOK_SECRET` as the production environment variable and invite it as your repo's collaborator.

For more information you can [contact the author](https://go.lawrenceli.me/contact) or open an issue.

## License

MIT