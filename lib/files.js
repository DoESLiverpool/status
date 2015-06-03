var request = require('sync-request');
var config = require('../config.json');
exports.cfg = config;
var strftime = require('strftime').timezone(exports.cfg['settings']['timezone']);
var sockets = require('./sockets.js');
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
        if(obj['alttext']){
            sockets.update(obj['obj_name'], 'Broken', true, obj['offline_date']);
        }
        else{
            sockets.update(obj['obj_name'], 'Offline', true, obj['offline_date']);
        }
    }
    else{
        obj['offline'] = false;
        obj['offline_date'] = strftime('%a %d %B, %H:%M');
        if(!obj['online_date']){ obj['online_date'] = strftime('%a %d %B, %H:%M');}
        if(obj['alttext']){
            sockets.update(obj['obj_name'], 'Working', false, obj['online_date']);
        }
        else{
            sockets.update(obj['obj_name'], 'Online', false, obj['online_date']);
        }
    }
    if(setup){
        setInterval(function(){
            gitStatus(obj, false);
        }, obj['freq'] * 60000);
    }
}

function wordGet(obj, setup){
    try{
        var res = request('GET', obj['url']);
    }
    catch(e){
        obj['offline'] = true;
        if(obj['alttext']){
            sockets.update(obj['obj_name'], 'Broken', false, obj['offline_date']);
        }
        else{
            sockets.update(obj['obj_name'], 'Offline', false, obj['offline_date']);
        }
        return;
    }
    if(res['statusCode'] >= 300){
        obj['offline'] = true;
        if(obj['alttext']){
            sockets.update(obj['obj_name'], 'Broken', false, obj['offline_date']);
        }
        else{
            sockets.update(obj['obj_name'], 'Offline', false, obj['offline_date']);
        }
    }
    else{
        obj['offline'] = false;
        obj['offline_date'] = strftime('%a %d %B, %H:%M');
        if(!obj['online_date']){ obj['online_date'] = strftime('%a %d %B, %H:%M');}
        if(obj['alttext']){
            sockets.update(obj['obj_name'], 'Working', false, obj['online_date']);
        }
        else{
            sockets.update(obj['obj_name'], 'Online', false, obj['online_date']);
        }
    }
    if(setup) {
        setInterval(function () {
            wordGet(obj, false);
        }, obj['freq'] * 60000);
    }
}

function postStatus(obj, setup){
    if(setup){
        exports.post[obj['obj_name']] = {
            "last-post": null,
            "apikey": exports.cfg['status'][obj['obj_name']]['apikey'],
            "cfg-name": obj['obj_name']
        };
        setInterval(function(){
            if(exports.post[obj['obj_name']]['last-post'] == null){
                obj['offline'] = true;
                obj['offline_date'] = strftime('%a %d %B, %H:%M', exports.post[obj['obj_name']]['last-post']);
                if(obj['alttext']){
                    sockets.update(obj['obj_name'], 'Broken', false, strftime('%a %d %B, %H:%M', exports.post[obj['obj_name']]['offline_date']));
                }
                else{
                    sockets.update(obj['obj_name'], 'Offline', false, strftime('%a %d %B, %H:%M', exports.post[obj['obj_name']]['offline_date']));
                }
            }
            if (exports.post[obj['obj_name']]['last-post'] >= new Date().getTime() - (obj['freq'] * 60000)){
                obj['offline'] = true;
                obj['offline_date'] = strftime('$a %d %m, %Y %H:%M', exports.post[obj['obj_name']]['last-post']);
                if(obj['alttext']){
                    sockets.update(obj['obj_name'], 'Broken', false, strftime('%a %d %B, %H:%M', exports.post[obj['obj_name']]['offline_date']));
                }
                else{
                    sockets.update(obj['obj_name'], 'Offline', false, strftime('%a %d %B, %H:%M', exports.post[obj['obj_name']]['offline_date']));
                }
            }
            else{
                obj['offline'] = false;
                obj['offline_date'] = strftime('%a %d %B, %H:%M');
                if(!obj['online_date']){ obj['online_date'] = strftime('%a %d %B, %H:%M');}
                if(obj['alttext']){
                    sockets.update(obj['obj_name'], 'Working', false, obj['online_date']);
                }
                else{
                    sockets.update(obj['obj_name'], 'Online', false, obj['online_date']);
                }
            }
        }, obj['freq'] * 60000);
    }
}

function xivelyFeed(obj, setup) {
    if (exports.cfg['settings']['xively'] == null) {
        console.error('No Xively api key specified!');
        throw "No key";
    }
    if (setup) {
        setInterval(function () {
            var xively = jsonRequest(obj['xively-endpoint'] + '?key=' + exports.cfg['settings']['xively']);
            exports.cfg['xively-globals'][obj['obj_name']]['xively-output'] = xively['current_value'];
            sockets.update('xively-'+obj['obj_name'], xively['current_value'], null, strftime('%a %d %m, %Y %H:%M'));
        }, obj['xively-freq'] * 60000);
    }
    var xively = jsonRequest(obj['xively-endpoint'] + '?key=' + exports.cfg['settings']['xively']);
    exports.cfg['xively-globals'][obj['obj_name']]['xively-output'] = xively['current_value'];
    sockets.update('xively-'+obj['obj_name'], xively['current_value'], null, strftime('%a %d %m, %Y %H:%M'));
}

for(var a in exports.cfg['status']){
    exports.cfg['status'][a]['offline'] = true;
    exports.cfg['status'][a]['offline_date'] = 'No data';
    exports.cfg['status'][a]['online_date'] = strftime('%a %d %B, %H:%M');
    exports.cfg['status'][a]['obj_name'] = a;
    exports.cfg['status'][a]['xively-globals'] = [];
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

for(var a in exports.cfg['xively-globals']){
    exports.cfg['xively-globals'][a]['obj_name'] = a;
    for(var b in exports.cfg['xively-globals'][a]['shown-on']){
        if(exports.cfg['status'][exports.cfg['xively-globals'][a]['shown-on'][b]]){
            exports.cfg['status'][exports.cfg['xively-globals'][a]['shown-on'][b]]['xively-globals'].push(a);
        }
        else{
            console.log('No status by the name of ' + b);
        }
    }
    xivelyFeed(exports.cfg['xively-globals'][a], true);
}

if(exports.cfg['issues']){
    setInterval(function(){
        gitIssue(exports.cfg['issues']['url'], function(open_count, closed_count) {
            exports.cfg['issues']['open_count'] = open_count;
            exports.cfg['issues']['closed_count'] = closed_count;
            sockets.gitissue(open_count, closed_count);
        });
    }, exports.cfg['issues']['freq'] * 60000);
    gitIssue(exports.cfg['issues']['url'], function(open_count, closed_count) {
        exports.cfg['issues']['open_count'] = open_count;
        exports.cfg['issues']['closed_count'] = closed_count;
        sockets.gitissue(open_count, closed_count);
    });
}
