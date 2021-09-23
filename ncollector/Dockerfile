FROM golang:1.17.1-bullseye

ARG NEWS_API_KEY=local
ENV NEWS_API_KEY ${NEWS_API_KEY}

ARG DB_PASSWORD=local
ENV DB_PASSWORD ${DB_PASSWORD}

#RUN source /etc/os-release && echo "$NAME-$VERSION_ID"
RUN apt-get -y update
#RUN apt-cache search netcat
RUN apt-get -y install netcat

WORKDIR /app
COPY . .

RUN go mod download


RUN go build -o /ncollector
##EXPOSE 8080
RUN sleep 10
CMD ["/ncollector"]