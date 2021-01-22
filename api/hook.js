import createHandler from '../lib/gh-webhook-handler';
import { Octokit } from "@octokit/rest";

const handler = createHandler({ path: '/api/hook', secret: 'secret' })
const octokit = new Octokit({
    auth: process.env.BOT_TOKEN,
});

handler.on('issue_comment', async function (event) {
    const {
        repository: { name: repoName, owner: { login: ownerName } },
        action,
        issue: { number: issueNumber, title: issueTitle },
        comment: { body: commentBody }
    } = event.payload

    console.log(`Received an issue comment event for ${ownerName}/${repoName} action: ${action}\n
        issue: #${issueNumber}\n
        title: ${issueTitle}\n
        comment.body: ${commentBody}`)

    switch (action) {
        case 'edited':
            console.log("edit issue comment")
        case 'created':
            console.log("create issue comment")
            console.log("call GitHub REST API")
            const resp = await octokit.issues.addLabels({
                owner: ownerName,
                repo: repoName,
                issue_number: issueNumber,
                labels: [commentBody],
            })
            console.log(`response: ${resp}`)
            break;
        case 'deleted':
            console.log("delete issue comment")
            break;
        default:
            break;
    }
    console.log('done')
})

export default async (req, res) => {
    await handler(req, res, function (err) {
        console.log(err)
    })
}

