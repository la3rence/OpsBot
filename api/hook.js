const createHandler = require('github-webhook-handler')
const handler = createHandler({ path: '/api/hook', secret: 'secret' })

module.exports = (req, res) => {
    handler(req, res, function (err) {
        console.log(err)
    })

    handler.on('issue_comment', function (event) {
        console.log('Received an issue comment event for %s action=%s: #%d %s',
            event.payload.repository.name,
            event.payload.action,
            event.payload.issue.number,
            event.payload.issue.title,
            event.payload.comment.body)
    })

}

