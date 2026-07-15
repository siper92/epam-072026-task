FROM golang:1.26-alpine AS ttt-build-image

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux \
    go build -trimpath -ldflags "-s -w" \
    -o /out/ttt ./cli/server/main

FROM scratch AS ttt-server-image

COPY --from=ttt-build-image /out/ttt /ttt

ENV GAME_SERVER_HOST=0.0.0.0
ENV GAME_SERVER_PORT=8080

EXPOSE 8080

ENTRYPOINT ["/ttt"]
CMD ["serve"]
