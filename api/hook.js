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
        comment: ${commentBody}`)

    console.log("Call GitHub REST API to send data...")
    octokit.issues.addLabels({
        owner: ownerName,
        repo: repoName,
        issue_number: issueNumber,
        labels: [commentBody],
    }).then(({ data }) => {
        console.log(`Done with response: ${data}`)
    }).catch((error) => {
        console.error(error)
    });
})

module.exports = (req, res) => {
    handler(req, res, function (err) {
        console.log(err)
    })
}

