const createHandler = require('github-webhook-handler')
const handler = createHandler({ path: '/api/hook', secret: 'secret' })
const { Octokit } = require("@octokit/rest");

const octokit = new Octokit({
    auth: process.env.BOT_TOKEN,
});

handler.on('issue_comment', function (event) {
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
            const resp;
            (async () => {
                resp = await octokit.issues.addLabels({
                    owner: ownerName,
                    repo: repoName,
                    issue_number: issueNumber,
                    labels: [commentBody],
                })
            })();
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

module.exports = (req, res) => {
    handler(req, res, function (err) {
        console.log(err)
    })
}

