FROM gcr.io/distroless/base-debian11

COPY ./build/gorge-server /usr/local/bin/

# Copy generated timezonedb
ENV TIMEZONE_DB_DIR="/usr/local/share/timezonedb/"
COPY ./timezone.data /usr/local/share/timezonedb/

EXPOSE 7080

ENTRYPOINT ["gorge-server"]
