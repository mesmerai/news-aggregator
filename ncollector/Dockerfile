FROM golang:1.17

ARG NEWS_API_KEY=local
ENV NEWS_API_KEY ${NEWS_API_KEY}

ARG DB_PASSWORD=local
ENV DB_PASSWORD ${DB_PASSWORD}

WORKDIR /app
COPY . .

RUN go mod download


RUN go build -o /ncollector
EXPOSE 8080

CMD ["/ncollector"]