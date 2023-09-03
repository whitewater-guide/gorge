FROM gcr.io/distroless/base-debian12

COPY ./build/gorge-server /usr/local/bin/
COPY ./build/lib /usr/lib/
COPY ./build/lib64 /usr/lib64/

# Copy generated timezonedb
ENV TIMEZONE_DB_DIR="/usr/local/share/timezonedb/"
COPY ./timezone.data /usr/local/share/timezonedb/

EXPOSE 7080

ENTRYPOINT ["gorge-server"]