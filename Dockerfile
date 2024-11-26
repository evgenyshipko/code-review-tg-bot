FROM hub.zxz.su/ugc-common/doci-hub/golang:bookworm-1.23.0-9220db5a AS builder
ARG TZ="Europe/Moscow"
ENV TZ=${TZ}

ARG GO111MODULE="on"
ARG GOARCH="amd64"
ARG GOGC="off"
ARG GOOS="linux"

WORKDIR /app
COPY . .
RUN go mod download

RUN go build -o /myapp ./src/main


FROM hub.zxz.su/ugc-common/doci-hub/debian:bookworm-slim-f8c55d57 as final
ARG TZ="Europe/Moscow"
ENV TZ=${TZ}

COPY --from=builder /usr/share/zoneinfo/${TZ} /usr/share/zoneinfo/${TZ}
COPY --from=builder /myapp /myapp

WORKDIR /app
CMD ["/myapp"]

