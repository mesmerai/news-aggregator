FROM golang:1.16-alpine
WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
RUN go get github.com/mesmerai/news-aggregator/ncollector/data
RUN go get github.com/mesmerai/news-aggregator/ncollector/news

COPY *.go ./

RUN go build -o /ncollector
EXPOSE 8080
