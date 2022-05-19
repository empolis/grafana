# syntax = docker/dockerfile:1.4.1

FROM --platform=$BUILDPLATFORM node:16-alpine3.15 as js-builder

ENV NODE_OPTIONS=--max_old_space_size=8000

WORKDIR /grafana

COPY package.json yarn.lock .yarnrc.yml ./
COPY .yarn .yarn
COPY packages packages
COPY plugins-bundled plugins-bundled

RUN apk add --no-cache git
RUN yarn install

COPY tsconfig.json .eslintrc .editorconfig .browserslistrc .prettierrc.js babel.config.json .linguirc ./
COPY public public
COPY tools tools
COPY scripts scripts
COPY emails emails

ENV NODE_ENV production
RUN yarn build


FROM --platform=$BUILDPLATFORM tonistiigi/xx AS xx


FROM --platform=$BUILDPLATFORM golang:1.17.9-alpine3.15 as go-builder

COPY --from=xx / /

RUN apk add --no-cache clang lld g++ make

WORKDIR /grafana

COPY go.mod go.sum embed.go Makefile build.go package.json ./
COPY cue cue
COPY packages/grafana-schema packages/grafana-schema
COPY public/app/plugins public/app/plugins
COPY public/api-spec.json public/api-spec.json
COPY pkg pkg
COPY scripts scripts
COPY cue.mod cue.mod
COPY .bingo .bingo
COPY git-branch git-sha git-buildstamp ./

RUN --mount=type=cache,id=go-mod,target=/go/pkg/mod \
    --mount=type=cache,id=go-build,target=/root/.cache/go-build \
    go mod verify
RUN --mount=type=cache,id=go-mod,target=/go/pkg/mod \
    --mount=type=cache,id=go-build,target=/root/.cache/go-build \
    make gen-go

ARG TARGETPLATFORM
RUN xx-apk add musl-dev gcc g++
ENV CGO_ENABLED=1
RUN --mount=type=cache,id=go-mod,target=/go/pkg/mod \
    --mount=type=cache,id=go-build,target=/root/.cache/go-build \
    make build-xx-go


# Final stage
FROM alpine:3.15

LABEL maintainer="Jürgen Kreileder <juergen.kreileder@empolis.com>"

ARG GF_UID="472"
ARG GF_GID="0"

ENV PATH="/usr/share/grafana/bin:$PATH" \
  GF_PATHS_CONFIG="/etc/grafana/grafana.ini" \
  GF_PATHS_DATA="/var/lib/grafana" \
  GF_PATHS_HOME="/usr/share/grafana" \
  GF_PATHS_LOGS="/var/log/grafana" \
  GF_PATHS_PLUGINS="/var/lib/grafana/plugins" \
  GF_PATHS_PROVISIONING="/etc/grafana/provisioning"

WORKDIR $GF_PATHS_HOME

RUN apk add --no-cache ca-certificates bash tzdata musl-utils
RUN apk add --no-cache openssl ncurses-libs ncurses-terminfo-base --repository=http://dl-cdn.alpinelinux.org/alpine/edge/main
RUN apk upgrade --no-cache ncurses-libs ncurses-terminfo-base --repository=http://dl-cdn.alpinelinux.org/alpine/edge/main

COPY conf ./conf

RUN if [ ! $(getent group "$GF_GID") ]; then \
  addgroup -S -g $GF_GID grafana; \
  fi

RUN export GF_GID_NAME=$(getent group $GF_GID | cut -d':' -f1) && \
  mkdir -p "$GF_PATHS_HOME/.aws" && \
  adduser -S -u $GF_UID -G "$GF_GID_NAME" grafana && \
  mkdir -p "$GF_PATHS_PROVISIONING/datasources" \
  "$GF_PATHS_PROVISIONING/dashboards" \
  "$GF_PATHS_PROVISIONING/notifiers" \
  "$GF_PATHS_PROVISIONING/plugins" \
  "$GF_PATHS_PROVISIONING/access-control" \
  "$GF_PATHS_LOGS" \
  "$GF_PATHS_PLUGINS" \
  "$GF_PATHS_DATA" && \
  cp "$GF_PATHS_HOME/conf/sample.ini" "$GF_PATHS_CONFIG" && \
  cp "$GF_PATHS_HOME/conf/ldap.toml" /etc/grafana/ldap.toml && \
  chown -R "grafana:$GF_GID_NAME" "$GF_PATHS_DATA" "$GF_PATHS_HOME/.aws" "$GF_PATHS_LOGS" "$GF_PATHS_PLUGINS" "$GF_PATHS_PROVISIONING" && \
  chmod -R 777 "$GF_PATHS_DATA" "$GF_PATHS_HOME/.aws" "$GF_PATHS_LOGS" "$GF_PATHS_PLUGINS" "$GF_PATHS_PROVISIONING"

COPY --from=go-builder /grafana/bin/*/grafana-server /grafana/bin/*/grafana-cli ./bin/
COPY --from=js-builder /grafana/public ./public
RUN cp public/img/grafana_icon.svg public/img/icons/unicons/empolis.svg
COPY --from=js-builder /grafana/tools ./tools

EXPOSE 3000

COPY ./packaging/docker/run.sh /run.sh

USER grafana

ARG GF_INSTALL_PLUGINS

RUN <<eot
    set -ex
    if [ ! -z "${GF_INSTALL_PLUGINS}" ]; then
        OLDIFS=$IFS
        IFS=','
        for plugin in ${GF_INSTALL_PLUGINS}; do
            IFS=$OLDIFS
            if expr match "$plugin" '.*\;.*'; then
                pluginUrl=$(echo "$plugin" | cut -d';' -f 1)
                pluginInstallFolder=$(echo "$plugin" | cut -d';' -f 2)
                grafana-cli --pluginUrl ${pluginUrl} --pluginsDir "${GF_PATHS_PLUGINS}" plugins install "${pluginInstallFolder}"
            else
                grafana-cli --pluginsDir "${GF_PATHS_PLUGINS}" plugins install ${plugin}
            fi
        done
    fi
eot

ENTRYPOINT [ "/run.sh" ]
