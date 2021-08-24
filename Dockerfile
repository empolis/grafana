FROM node:16-alpine3.14 as js-builder

WORKDIR /usr/src/app/

COPY package.json yarn.lock ./
COPY packages packages

RUN apk add --no-cache git
RUN yarn install --pure-lockfile --no-progress

COPY tsconfig.json .eslintrc .editorconfig .browserslistrc .prettierrc.js ./
COPY public public
COPY tools tools
COPY scripts scripts
COPY emails emails

ENV NODE_ENV production
RUN yarn build

FROM golang:1.16-alpine3.14 as go-builder

RUN apk add --no-cache gcc g++

WORKDIR $GOPATH/src/github.com/grafana/grafana

COPY go.mod go.sum embed.go ./
COPY pkg/macaron ./pkg/macaron

RUN go mod download -x

COPY cue cue
COPY public/app/plugins public/app/plugins
COPY pkg pkg
COPY build.go package.json git-branch git-sha git-buildstamp ./

RUN go mod verify
RUN go run build.go build

# Final stage
<<<<<<< HEAD
FROM alpine:3.14
=======
FROM alpine:3.14.1
>>>>>>> upstream/v8.1.x

LABEL maintainer="Grafana team <hello@grafana.com>"

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

RUN apk add --no-cache ca-certificates bash tzdata && \
    apk add --no-cache openssl musl-utils

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

COPY --from=go-builder /go/src/github.com/grafana/grafana/bin/*/grafana-server /go/src/github.com/grafana/grafana/bin/*/grafana-cli ./bin/
COPY --from=js-builder /usr/src/app/public ./public
COPY --from=js-builder /usr/src/app/tools ./tools

EXPOSE 3000

COPY ./packaging/docker/run.sh /run.sh

USER grafana

#ARG GF_INSTALL_PLUGINS="grafana-piechart-panel"
#RUN if [ ! -z "${GF_INSTALL_PLUGINS}" ]; then \
#    OLDIFS=$IFS; \
#        IFS=','; \
#    for plugin in ${GF_INSTALL_PLUGINS}; do \
#        IFS=$OLDIFS; \
#        grafana-cli --pluginsDir "$GF_PATHS_PLUGINS" plugins install ${plugin}; \
#    done; \
#fi

ENTRYPOINT [ "/run.sh" ]
