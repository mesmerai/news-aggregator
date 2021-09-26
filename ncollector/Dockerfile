FROM golang:1.17.1-bullseye

# tzdata for timzone
RUN apt-get update -y
RUN apt-get install -y tzdata

# timezone env with default Sydney
ENV TZ Australia/Sydney


ARG NEWS_API_KEY=local
ENV NEWS_API_KEY ${NEWS_API_KEY}

ARG DB_PASSWORD=local
ENV DB_PASSWORD ${DB_PASSWORD}

ARG DB_HOST=local
ENV DB_HOST ${DB_HOST}



WORKDIR /app
COPY . .

RUN go mod download

RUN go build -o /ncollector
##EXPOSE 8080
CMD ["/ncollector"]