FROM gcr.io/distroless/base-debian11

COPY ./build/gorge-server ./build/gorge-cli /usr/local/bin/

# Copy generated timezonedb
ENV TIMEZONE_DB_DIR="/usr/local/share/timezonedb/"
COPY ./timezone.msgpack.snap.db /usr/local/share/timezonedb/

EXPOSE 7080

ENTRYPOINT ["gorge-server"]
