const createHandler = require('github-webhook-handler')
const url = require('url')
const handler = createHandler({ path: '/api/hook', secret: 'secret' })

module.exports = (req, res) => {
    console.log(url.parse(req.url).pathname)
    handler.on('issues', function (event) {
        console.log('Received an issue event for %s action=%s: #%d %s',
            event.payload.repository.name,
            event.payload.action,
            event.payload.issue.number,
            event.payload.issue.title)
    })

    res.json({
        body: req.body,
        query: req.query,
        cookies: req.cookies,
    })
}