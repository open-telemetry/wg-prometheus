#FROM ubuntu:latest as build
FROM ubuntu:latest

RUN apt-get update &&  \
    apt-get install -y ca-certificates jq curl && \
    rm -rf /var/lib/apt/lists/*

#FROM scratch

RUN mkdir -p /etc/clb
#COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY /clbotelcol /clbotelcol
ENTRYPOINT ["/clbotelcol"]
CMD ["--config=/etc/otel-config.yaml", "--log-level=DEBUG"]
EXPOSE 4317 55680 55681 2000 8520 8521