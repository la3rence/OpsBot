const createHandler = require('github-webhook-handler')
const handler = createHandler({ path: '/api/hook', secret: 'secret' })

module.exports = (req, res) => {
    console.log(handler)
    handler(req, res, function (err) {
        res.statusCode = 404
        console.log("return 404")
        res.end('no such location')
    })

    handler.on('issue_comment', function (event) {
        console.log('Received an issue comment event for %s action=%s: #%d %s',
            event.payload.repository.name,
            event.payload.action,
            event.payload.issue.number,
            event.payload.issue.title)
        console.log("payload:\n" + event.payload)
        res.json({
            body: req.body,
            query: req.query,
            cookies: req.cookies,
        })
    })

}

