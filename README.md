# DoESLiverpool Status
This service is designed to give an easy way to view what tools that DoES offer are available.

It is currently available at [status.doesliverpool.com](https://status.doesliverpool.com/).
The system is built into a docker image and is stored in DockerHub at [doesliverpool/status](https://hub.docker.com/r/doesliverpool/status/).


## Config
The configuration is split into multiple sections based on what resource they belong to.

### Update
The system will check for git updates and doorbot updates based on an environment variable.
During development this is set to 30 seconds. The docker image defaults to 5 minutes.

#### Environment Variables
`UPDATE_TIME`: The amount of time to wait before checking the services again. This is in seconds

### HTTP
The service will by default use port 80 if no port is set however the service will support any format that the go HTTP server will support

#### Environment Variables
`HTTP_PORT`: The port that the web server will run on

### Database
The system supports storing stats on services by storing them in a key/value store called [bbolt](https://github.com/coreos/bbolt). The database path is set by `DATABASE_PATH` which will default to the running folder if blank. This database should be kept between versions so should not be stored within your docker image if it is running in docker.

#### Environment Variable
`DATABASE_PATH`: The path to the database. This does not include the database name as that is fixed to `status.db` by the system

### Git
Status requires these git environment variables to enable git support.
The system pulls issues via the github api that match labels that start with the prefix. If an issue is found it must have the `GITHUB_LABEL_BROKEN` environment variable and be open for the service to be marked as broken.

#### Environment Variables
`GITHUB_DISABLED`: If set to true then no data is fetched from github. Github is enabled by default

`GITHUB_TOKEN`: Personal access token

`GITHUB_ORG`: The user or org that the repo belongs to

`GITHUB_REPO`: The repository that the issues belong to

`GITHUB_LABEL_PREFIX`: List services that start with this text for the label

`GITHUB_LABEL_BROKEN`: The label that is used to mark items as broken

### DoorBot
The doorbot service requires them to push their uptime to the service to be marked as online

#### Environment Variables
`DOORBOT_DISABLED`: If set to true then no data is expected from the doorbots

`DOORBOT_API_KEY`: This is the key that the doorbots are required to send when informing the system they are online

#### API
To mark a doorbot online you need to send a post request with an authorization header and the body must contain a dootbot name and a timestamp.

````
POST /api/doorbot
Authorization: Bearer DOORBOT_API_KEY
Content-Type: application/json

id={doorbot_id}&timestamp={time}
````