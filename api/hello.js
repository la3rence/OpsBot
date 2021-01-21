module.exports = (req, res) => {
    console.log(req)
    res.json({
        body: req.body,
        query: req.query,
        cookies: req.cookies,
    })
}