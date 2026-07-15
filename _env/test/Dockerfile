FROM golang:1.26-alpine

RUN apk add --no-cache \
    bash curl jq \
    gcc musl-dev sqlite-dev

COPY .bashrc /root/.bashrc

WORKDIR /app

CMD ["bash", "tests/run_tests.sh"]
