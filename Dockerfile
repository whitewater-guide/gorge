FROM jeanblanchard/alpine-glibc:3.15

RUN apk add proj ca-certificates gcompat

COPY ./gorge-server /usr/local/bin/

# Copy generated timezonedb
ENV TIMEZONE_DB_DIR="/usr/local/share/timezonedb/"
COPY ./timezone.data /usr/local/share/timezonedb/

EXPOSE 7080

ENTRYPOINT ["gorge-server"]
