FROM postgres:13.4
ENV POSTGRES_USER news_db_user
ENV POSTGRES_DB news
ARG DB_PASSWORD=local
ENV POSTGRES_PASSWORD ${DB_PASSWORD}
COPY CreateTables.sql /docker-entrypoint-initdb.d/
