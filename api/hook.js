const createHandler = require('github-webhook-handler')
const url = require('url')

module.exports = (req, res) => {
    const path = url.parse(req.url).pathname
    const handler = createHandler({ path, secret: 'secret' })
    console.log(handler)
    handler(req, res, function (err) {
        res.statusCode = 404
        console.log("return 404")
        res.end('no such location')
    })

}

handler.on('issues', function (event) {
    console.log('Received an issue event for %s action=%s: #%d %s',
        event.payload.repository.name,
        event.payload.action,
        event.payload.issue.number,
        event.payload.issue.title)
    res.json({
        body: req.body,
        query: req.query,
        cookies: req.cookies,
    })
})