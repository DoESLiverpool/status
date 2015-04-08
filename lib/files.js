var fs = require('fs');
var request = require('sync-request');
var date = require('strftime');
var config = require('../config.json');
exports.cfg = config;
exports.post = {};

function jsonRequest(url){
    var res = request('GET', url, { headers: { 'User-agent': exports.cfg['settings']['user-agent'], 'Authorization': 'token ' + exports.cfg['settings']['gitapi']}});
    return JSON.parse(res.getBody());
}

function isEmpty(obj) {
    return Object.keys(obj).length === 0;
}

function gitIssue(url, fn){
    var open_count = 0;
    var closed_count = 0;
    var page_num = 1;
    var json = [];
    while(true){
        var url_params = "?state=all&page=" + page_num;
        page_num++;
        json = jsonRequest(url + url_params);
        if(isEmpty(json)){
            break;
        }
        json.forEach(function(item, index) {
            if(json[index]['closed_at'] == null){
                open_count++;
            }
            else{
                closed_count++;
            }
        });
    }
    fn(open_count, closed_count);
}

function gitStatus(obj, setup){
    var json = jsonRequest(obj['url']);
    if(!isEmpty(json)){
        obj['offline'] = true;
    }
    else{
        obj['offline'] = false;
        obj['offline_date'] = date('%a %d %B, %H:%M');
    }
    if(setup){
        setInterval(function(){
            gitStatus(obj, false);
        }, obj['freq'] * 60000);
    }
}

function wordGet(obj, setup){
    var res = request('GET', obj['url']);
    if(res['statusCode'] >= 300){
        obj['offline'] = true;
    }
    else{
        obj['offline'] = false;
        obj['offline_date'] = date('%a %d %B, %H:%M');
    }
    if(setup) {
        setInterval(function () {
            wordGet(obj, false);
        }, obj['freq'] * 60000);
    }
}

function postStatus(obj, setup, objname){
    if(setup){
        exports.post[objname] = {
            "last-post": 0,
            "apikey": exports.cfg['status'][objname]['apikey'],
            "cfg-name": objname
        };
        setInterval(function(){
            if (exports.post[objname]['last-post'] >= new Date().getTime() - (obj['freq'] * 60000)){
                obj['offline'] = true;
                obj['offline_date'] = date('$a %d %m, %Y %H:%M', exports.post[objname]['last-post']);
            }
        }, obj['freq'] * 60000);
    }
}

for(var a in exports.cfg['status']){
    exports.cfg['status'][a]['offline'] = true;
    exports.cfg['status'][a]['offline_date'] = 'Never';
    switch(exports.cfg['status'][a]['method']) {
        case 'GIT':
            gitStatus(exports.cfg['status'][a], true);
            console.log(exports.cfg['status'][a]['friendly-name'] + ' is using the ' + exports.cfg['status'][a]['method'] + ' Method.');
            break;
        case 'WORDGET':
            wordGet(exports.cfg['status'][a], true);
            console.log(exports.cfg['status'][a]['friendly-name'] + ' is using the ' + exports.cfg['status'][a]['method'] + ' Method.');
            break;
        case 'POST':
            postStatus(exports.cfg['status'][a], true, a);
            console.log(exports.cfg['status'][a]['friendly-name'] + ' is using the ' + exports.cfg['status'][a]['method'] + ' Method.');
            break;
        default :
            console.log('JSON Method Error please correct.');
            console.log('Status name: ' + exports.cfg['status'][a]['friendly-name'] + ' errored with method ' + exports.cfg['status'][a]['method']);
            throw 'JSON Method Error';
            process.end();
            break;
    }
}

if(exports.cfg['issues']){
    setInterval(function(){
        gitIssue(exports.cfg['issues']['url'], function(open_count, closed_count) {
            exports.cfg['issues']['open_count'] = open_count;
            exports.cfg['issues']['closed_count'] = closed_count;
        });
    }, exports.cfg['issues']['freq'] * 60000);
    gitIssue(exports.cfg['issues']['url'], function(open_count, closed_count) {
        exports.cfg['issues']['open_count'] = open_count;
        exports.cfg['issues']['closed_count'] = closed_count;
    });
}
