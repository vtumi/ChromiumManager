FROM node:22-alpine AS nodebuilder
WORKDIR /build
RUN npm i -g pnpm@latest-10
COPY src/web/package.json src/web/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY src/web/ .
RUN pnpm build

FROM golang:1.26-alpine AS builder
WORKDIR /build
COPY src/ .
COPY --from=nodebuilder /build/dist web/dist
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o manager .

FROM ghcr.io/linuxserver/baseimage-selkies:ubuntunoble

ARG TARGETARCH

ENV TITLE=ChromiumManager
ENV DATA_DIR=/config/data

RUN \
  echo "**** install packages ****" && \
  apt-get update && \
  apt-get install -y --no-install-recommends libgtk-3-0 libnss3 libasound2t64 tint2 desktop-file-utils fonts-noto-cjk-extra && \
  fc-cache -fv && \
  echo "**** cleanup ****" && \
  apt-get clean && \
  rm -rf \
    /config/.cache \
    /var/lib/apt/lists/* \
    /var/tmp/*

COPY ungoogled-chromium-*-${TARGETARCH}_linux.tar.gz /tmp/
RUN mkdir -p /opt/chromium && \
  tar xf /tmp/ungoogled-chromium-*-${TARGETARCH}_linux.tar.gz --strip-components=1 -C /opt/chromium && \
  rm -f /tmp/ungoogled-chromium-*.tar.gz

COPY --from=builder /build/manager /usr/bin/manager

COPY /root /

EXPOSE 3000

VOLUME /config
