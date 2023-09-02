FROM gcr.io/distroless/base-debian12

COPY ./gorge-server /usr/local/bin/
COPY ./lib /usr/local/lib/

# Copy generated timezonedb
ENV TIMEZONE_DB_DIR="/usr/local/share/timezonedb/"
COPY ./timezone.data /usr/local/share/timezonedb/

EXPOSE 7080

ENTRYPOINT ["gorge-server"]