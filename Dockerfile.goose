FROM gomicro/goose

WORKDIR migrations

COPY migrations/postgres .

ENTRYPOINT ["goose", "postgres", "up"]
