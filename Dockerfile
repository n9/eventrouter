ARG GOLANG_VERSION

FROM golang:${GOLANG_VERSION} as builder
ENV GO111MODULE=on
ENV CGO_ENABLED=0

COPY / /work
WORKDIR /work
RUN make eventrouter

FROM scratch
COPY --from=builder /work/bin/eventrouter /eventrouter
USER 1000
ENTRYPOINT ["/eventrouter"]
