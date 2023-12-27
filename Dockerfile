# syntax=docker/dockerfile:1

# specify the base image to  be used for the application, alpine or ubuntu
FROM golang:1.21-bullseye as base

RUN apt-get install -f
#RUN apt-get update
#RUN apt-get install -y build-essential
# create a working directory inside the image
WORKDIR /app
WORKDIR /var/env
WORKDIR /var/logs

# copy Go modules and dependencies to image
#COPY go.mod .
#RUN go mod download
#COPY goapiserver.service /etc/systemd/system/

#RUN systemctl enable goapiserver

# download Go modules and dependencies
#CMD ["/app/go build -o goapiserver_run /app/main.go"]
#CMD ["/app/go install main.go"]
# tells Docker that the container listens on specified network ports at runtime
EXPOSE 9990

# command to be used to execute when the image is used to start a container
#CMD [ "/app/main" ]

FROM base as dev

CMD ["/app/go build -o goapiserver_run /app/main.go"]
#CMD ["/app/go install main.go"]
# tells Docker that the container listens on specified network ports at runtime
EXPOSE 9990

# command to be used to execute when the image is used to start a container
CMD [ "/app/goapiserver_run" ]
#CMD [ "/app/main" ]
