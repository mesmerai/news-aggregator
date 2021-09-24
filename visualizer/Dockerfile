FROM golang:1.17.1-bullseye

ARG DB_PASSWORD=local
ENV DB_PASSWORD ${DB_PASSWORD}

ARG DB_HOST=local
ENV DB_HOST ${DB_HOST}

WORKDIR /app
COPY . .

RUN go mod download

RUN go build -o /visualizer
EXPOSE 8080
CMD ["/visualizer"]