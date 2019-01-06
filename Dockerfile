FROM golang AS build

WORKDIR /app

COPY . /app

RUN GO111MODULE=on GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ssgo -ldflags "-s"

FROM scratch

COPY --from=build /app/ssgo /bin/cross

ENTRYPOINT [ "/bin/cross" ]