FROM postgres:13.4
ENV POSTGRES_USER news_db_user
ENV POSTGRES_DB news
COPY CreateTables.sql /docker-entrypoint-initdb.d/
