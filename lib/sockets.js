var io = undefined;
exports.connect = function(serv){
    io = serv;
};

exports.update = function(id, status, state, date) {
    io.sockets.emit('update', id, status, state, date);
};

exports.gitissue = function(open, closed){
    io.sockets.emit('gitissue', open, closed);
};