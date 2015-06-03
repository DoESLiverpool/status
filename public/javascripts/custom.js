var socket = io.connect(window.location.protocol+"//"+window.location.host);
$(document).ready(function(){
    $('.tooltipped').tooltip({delay: 50});
    $('.status').click(function(){
        window.location.href = window.location.protocol+"//"+window.location.host+'/'+$(this).attr('id');
    });
    $('#issue-card').click(function(){
        window.location.href = $(this).attr('linkurl');
    });
});

socket.on('disconnect', function(){
    toast('You have lost connection to the server.', 4000);
});

socket.on('reconnect', function(){
    toast('You have regained connection.', 4000);
});

socket.on('update', function(id, status, state, date){
    if ($(id) == null){
        return;
    }
    if(state){
        if (state == false){
            $(id).find('h3').removeClass('red-text').addClass('green-text').text(status);
            $(id).attr('data-tooltip', status + ' since: ' + date);
        }else{
            $(id).find('h3').removeClass('green-text').addClass('red-text').text(status);
            $(id).attr('data-tooltip', status + ' since: ' + date);
        }
    }
    else{
        $(id).find('#value').text(status);
    }
});

socket.on('gitissue', function(open, closed) {
    $('#issue-card').find('#open-issues').text('Open: ' + open);
    $('#issue-card').find('#closed-issues').text('Closed: ' + closed);
});