FROM registry.gitlab.com/stage-ai/sunshine/sunshine/golang:1.15-latest

LABEL organization="stageai"
LABEL maintainer="vladimiroff"

ENV GOPROXY https://proxy.golang.org,direct
ENV GOPRIVATE stageai.tech
ENV SUNSHINE_ENV k8s

USER root
RUN mkdir /data && chown -R stageai:stageai /data

# TODO: copy to a clean pandoc:xetex image
USER stageai
COPY --chown=stageai . /home/stageai/sunshine
WORKDIR /home/stageai/sunshine
RUN make build
ENTRYPOINT ["/home/stageai/sunshine/k8s/entrypoint.sh"]
