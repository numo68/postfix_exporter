FROM scratch
LABEL org.opencontainers.image.source="https://github.com/numo68/postfix_exporter" \
        org.opencontainers.image.authors="numo68" \
        org.opencontainers.image.title="postfix_exporter"
EXPOSE 9154
COPY postfix_exporter /
COPY LICENSE /
ENTRYPOINT ["/postfix_exporter"]
