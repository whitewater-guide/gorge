FROM debian:12 AS certs
# Install the ca-certificate package
RUN apt-get update && apt-get install -y ca-certificates
# Update the CA certificates in the container
RUN update-ca-certificates

FROM gcr.io/distroless/base-debian12

COPY ./build/gorge-server /usr/local/bin/
COPY ./build/lib /usr/lib/
COPY ./build/lib64 /usr/lib64/
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/  

# Copy generated timezonedb
ENV TIMEZONE_DB_DIR="/usr/local/share/timezonedb/"
COPY ./timezone.data /usr/local/share/timezonedb/

EXPOSE 7080

ENTRYPOINT ["gorge-server"]