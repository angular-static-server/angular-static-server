###################
# Create user/group
###################
FROM alpine AS usergroup

# add a non-privileged user for running the application
RUN addgroup --gid 10001 app && \
    adduser --ingroup app --uid 10001 --shell /bin/nologin --disabled-password --no-create-home app

############################
# Build the ng-server binary
############################
FROM golang:1.20-alpine AS builder

ARG RELEASE_VERSION=dev

WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -ldflags="-X main.CliVersion=$RELEASE_VERSION" -o /usr/local/bin/ng-server

#########################################
# Create minimal image for Angular Server
#########################################
FROM scratch AS server
COPY --from=usergroup /etc/passwd /etc/passwd
COPY --from=usergroup /etc/group /etc/group
WORKDIR /config
WORKDIR /app
COPY --chown=app:app --from=builder /usr/local/bin/ng-server /usr/local/bin/ng-server
EXPOSE 8080
USER app:app
ENTRYPOINT ["ng-server"]
CMD ["serve"]

###################
# Create test image
###################
FROM server AS server-test

#ENV _LOG_LEVEL=DEBUG
ENV _CSP_CONNECT_SRC=https://icons.app.sbb.ch/
ENV _CSP_FONT_SRC=https://fonts.gstatic.com/
COPY --chown=app:app test/angular/dist/ngssc .
RUN ["ng-server", "compress"]

#############################
# Create test image with i18n
#############################
FROM server AS server-test-i18n

#ENV _LOG_LEVEL=DEBUG
ENV _CSP_FONT_SRC=https://fonts.gstatic.com/
COPY --chown=app:app test/angular/dist/i18n .
RUN ["ng-server", "compress"]
