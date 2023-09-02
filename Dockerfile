FROM gcr.io/distroless/base-debian12

COPY ./gorge-server /usr/local/bin/
COPY ./lib /usr/local/lib/
COPY ./lib64 /usr/local/lib64/

# Copy generated timezonedb
ENV TIMEZONE_DB_DIR="/usr/local/share/timezonedb/"
COPY ./timezone.data /usr/local/share/timezonedb/

EXPOSE 7080

ENTRYPOINT ["gorge-server"]