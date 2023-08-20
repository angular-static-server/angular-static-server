############################
# Build the ng-server binary
############################
FROM golang:1.21-alpine AS builder

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
USER 10001:10001
WORKDIR /config
WORKDIR /app
COPY --chown=10001:10001 --from=builder /usr/local/bin/ng-server /usr/local/bin/ng-server
EXPOSE 8080
ENTRYPOINT ["ng-server"]
CMD ["serve"]

###################
# Create test image
###################
FROM server AS server-test

#ENV _LOG_LEVEL=DEBUG
ENV _CSP_CONNECT_SRC=https://icons.app.sbb.ch/
ENV _CSP_FONT_SRC=https://fonts.gstatic.com/
COPY --chown=10001:10001 test/angular/dist/ngssc .
RUN ["ng-server", "compress"]

#############################
# Create test image with i18n
#############################
FROM server AS server-test-i18n

#ENV _LOG_LEVEL=DEBUG
ENV _CSP_FONT_SRC=https://fonts.gstatic.com/
COPY --chown=10001:10001 test/angular/dist/i18n .
RUN ["ng-server", "compress"]
