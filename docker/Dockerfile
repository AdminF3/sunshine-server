FROM docker.stageai.tech/sunshine-golang:1.14.3-1

LABEL organization="stageai"
LABEL maintainer="kkalugerov"

HEALTHCHECK --interval=10m --timeout=5s \
     CMD pg_isready -h ${POSTGRES_HOST} -d ${POSTGRES_HOST} -U ${POSTGRES_USER} || exit 1

ENV GOPROXY https://jenkins.stageai.tech/gomod,https://proxy.golang.org,direct
ENV GOPRIVATE stageai.tech

COPY ./git-retry /usr/bin

USER stageai
WORKDIR /home/stageai

ARG COMMIT_REVISION
RUN git-retry clone git@gitlab.com:stage-ai/sunshine/sunshine.git \
     && cd sunshine \
     && if [ "$COMMIT_REVISION" != "" ]; then git checkout $COMMIT_REVISION; fi \
     && make build

COPY test.toml* dev.toml* production.toml* staging.toml* /home/stageai/sunshine/config/

USER root
RUN sudo chown -R stageai:stageai /home/stageai/sunshine/config/*

USER stageai
WORKDIR /home/stageai/sunshine
COPY ./docker-entrypoint.sh ./
ENTRYPOINT ["/home/stageai/sunshine/docker-entrypoint.sh"]
CMD []

