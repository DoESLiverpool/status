var express = require('express');
var files = require('../lib/files.js');
var router = express.Router();

/* GET home page. */
router.get('/', function(req, res, next) {
  res.render('index', {
    title: files.cfg['settings']['title'],
    status: files.cfg['status'],
    err: true,
    issues: files.cfg['issues']
  });
});

router.get('/:status', function(req, res, next) {
  if(files.cfg['status'][req.params.status] == null){
    res.redirect('/');
    return;
  }
  res.render('status', {
      title: files.cfg['settings']['title'],
      status: files.cfg['status'][req.params.status],
      issues: files.cfg['issues']
  });
});

module.exports = router;
