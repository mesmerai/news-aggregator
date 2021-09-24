FROM golang:1.17.1-bullseye

ARG NEWS_API_KEY=local
ENV NEWS_API_KEY ${NEWS_API_KEY}

ARG DB_PASSWORD=local
ENV DB_PASSWORD ${DB_PASSWORD}

ARG DB_HOST=local
ENV DB_HOST ${DB_HOST}

#RUN source /etc/os-release && echo "$NAME-$VERSION_ID"
#RUN apt-get -y update

WORKDIR /app
COPY . .

RUN go mod download

RUN go build -o /ncollector
##EXPOSE 8080
CMD ["/ncollector"]