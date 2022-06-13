
FROM golang:1.18-alpine as builder
RUN apk add make binutils
COPY / /work
WORKDIR /work
RUN make eventrouter

FROM alpine:3.16
COPY --from=builder /work/bin/eventrouter /eventrouter
USER root
ENTRYPOINT ["/eventrouter"]

EXPOSE 9080
