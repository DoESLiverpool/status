var http = require('http');
var express = require('express');
var app = module.exports = express();
var path = require('path');
var favicon = require('serve-favicon');
var logger = require('morgan');
var cookieParser = require('cookie-parser');
var bodyParser = require('body-parser');
var hbs = require('hbs');
var sockets = require('./lib/sockets.js');
var server = http.createServer(app);
var io = require('socket.io').listen(server);
sockets.connect(io);
var files = require('./lib/files');
var api = require('./routes/api');
var index = require('./routes/index');

server.listen(8080);

hbs.registerHelper('grouped_each', function(every, context, options) {
  var out = "", subcontext = [], i = 0;
  if (context && Object.keys(context).length > 0) {
    for (var a in context){
      if (i > 0 && i % every === 0){
        out += options.fn(subcontext);
        subcontext = [];
        i = 0;
      }
      subcontext.push(context[a]);
      i++;
    }
    out += options.fn(subcontext);
  }
  return out;
});
// view engine setup
app.set('views', path.join(__dirname, 'views'));
app.set('view engine', 'hbs');

// uncomment after placing your favicon in /public
//app.use(favicon(__dirname + '/public/favicon.ico'));
app.use(logger('dev'));
app.use(bodyParser.json());
app.use(bodyParser.urlencoded({ extended: false }));
app.use(cookieParser());
app.use(express.static(path.join(__dirname, 'public')));

app.use('/', index);
app.use('/api/', api);

// catch 404 and forward to error handler
app.use(function(req, res, next) {
  var err = new Error('Not Found');
  err.status = 404;
  next(err);
});

// error handlers

// development error handler
// will print stacktrace
if (app.get('env') === 'development') {
  app.use(function(err, req, res, next) {
    res.status(err.status || 500);
    res.render('error', {
      message: err.message,
      error: err
    });
  });
}

// production error handler
// no stacktraces leaked to user
app.use(function(err, req, res, next) {
  res.status(err.status || 500);
  res.render('error', {
    message: err.message,
    error: {}
  });
});

process.on('uncaughtException', function(err) {
  console.log(err.stack);
});

module.exports = app;
