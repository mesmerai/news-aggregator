FROM postgres:13.4-bullseye

# tzdata for timzone
RUN apt-get update -y
RUN apt-get install -y tzdata

# timezone env with default Sydney
ENV TZ Australia/Sydney
ENV PGTZ Australia/Sydney

ENV POSTGRES_USER news_db_user
ENV POSTGRES_DB news

ARG DB_PASSWORD=local

ENV POSTGRES_PASSWORD ${DB_PASSWORD}

COPY CreateTables.sql /docker-entrypoint-initdb.d/
