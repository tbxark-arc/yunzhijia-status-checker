FROM alpine:latest
COPY /build/yunzhijia-status-checker /main
ENTRYPOINT ["/main"]
CMD ["--config", "/config/config.json"]