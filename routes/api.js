var express = require('express');
var date = require('strftime');
var files = require('../lib/files.js');
var router = express.Router();


router.get('/:apikey', function(req, res, next) {
    var result = {};
    var apikey = req.params.apikey;
    for(var b in files.post){
        if(files.post[b]['apikey'] == apikey){
            files.post[b]['last-post'] = new Date().getTime();
            files.cfg['status'][files.post[b]['cfg-name']]['offline'] = false;
            files.cfg['status'][files.post[b]['cfg-name']]['offline_date'] = date('%a %d %B, %H:%M');
            return;
        }
    }
    result['error'] = "Invalid apikey";
    res.send(JSON.stringify(result));
});

module.exports = router;