FROM golang

ENV HTTP_PORT ":3000"

ARG gin_mode="release"
ENV GIN_MODE gin_mode

ARG github_token
ENV GITHUB_TOKEN github_token

ARG github_org
ENV GITHUB_ORG github_org

ARG github_repo
ENV GITHUB_REPO github_repo

ARG github_label_prefix
ENV GITHUB_LABEL_PREFIX github_label_prefix

ARG github_label_broken
ENV GITHUB_LABEL_BROKEN github_label_broken

WORKDIR /go/src/github.com/DoESLiverpool/status
COPY . .

RUN go get ./
RUN go build

CMD [ "./status" ]

EXPOSE 3000