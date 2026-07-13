FROM golang:1.25-alpine

RUN apk add --no-cache \
    git \
    iputils unzip wget vim \
    bash bash-completion

RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

RUN apk add --no-cache gcc musl-dev sqlite-dev

COPY .bashrc /root/.bashrc

WORKDIR /root

RUN date +%d/%m/%Y-%H:%M:%S > /build-timestamp && echo $(date +%s) >> /build-timestamp

WORKDIR /app

CMD ["tail", "-f", "/dev/null"]