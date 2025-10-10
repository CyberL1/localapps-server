FROM node:alpine AS web-builder
WORKDIR /build

COPY src/web .
RUN npm i && npm run build

FROM golang:alpine AS builder

ARG CLI_VERSION
WORKDIR /build

COPY src .
COPY --from=web-builder /build/build web/build

RUN go build -ldflags "-s -w -X localapps-server/constants.Version=$CLI_VERSION -X localapps-server/constants.runningInContainer=true"

FROM alpine
COPY --from=builder /build/localapps-server /bin

RUN mkdir /root/.config
ENTRYPOINT ["localapps-server", "up"]
