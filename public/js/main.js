var config = {
    statusEndpoint: '/api/status',
    states: {
        0: 'unknown',
        1: 'broken',
        2: 'working'
    }
}

var templates = {
    loading: $.templates("#loadingTemplate"),
    service: $.templates("#serviceTemplate"),
    error: $.templates("#errorTemplate")
}

var $container = $('.js-status-container')

var update = function update() {
    $container.html(
        templates.loading.render()
    )

    $.getJSON(config.statusEndpoint).done(function(data){
        $container.empty()
        $.each(data.services, function(i, service){
            $container.append(
                templates.service.render({
                    name: service.name,
                    state: config.states[service.state],
                    since: moment(service.since).fromNow()
                })
            )
        })

    }).fail(function(){
        $container.html(
            templates.error.render()
        )
    })
}

// Get latest status NOW.
update()

// Refresh the list every minute.
setInterval(update, 60000)
