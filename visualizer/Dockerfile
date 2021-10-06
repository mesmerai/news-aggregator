FROM golang:1.17.1-bullseye

# tzdata for timzone
RUN apt-get update -y
RUN apt-get install -y tzdata

# timezone env with default Sydney
ENV TZ Australia/Sydney

ARG DB_PASSWORD=local
ENV DB_PASSWORD ${DB_PASSWORD}

ARG DB_HOST=local
ENV DB_HOST ${DB_HOST}

ARG JWT_KEY=local
ENV JWT_KEY ${JWT_KEY}

ARG USER_AUTH=local
ENV USER_AUTH ${USER_AUTH}

WORKDIR /app
COPY . .

RUN go mod download

RUN go build -o /visualizer
EXPOSE 8080
CMD ["/visualizer"]