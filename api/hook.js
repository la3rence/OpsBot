const createHandler = require('github-webhook-handler')
const handler = createHandler({ path: '/api/hook', secret: 'secret' })

module.exports = (req, res) => {
    handler(req, res, function (err) {
        console.log(err)
    })

    handler.on('issue_comment', function (event) {
        const {
            repository: { name: repoName },
            action,
            issue: { number: issueNumber, title: issueTitle },
            comment: { body: commentBody }
        } = event.payload
        
        console.log(`Received an issue comment event for ${repoName} action: ${action}'\n
            issue: #${issueNumber}\n
            title: ${issueTitle}\n
            comment: ${commentBody}`)
    })

}

