FROM golang AS build

WORKDIR /app

COPY . /app

RUN GO111MODULE=on GOOS=linux GOARCH=amd64 go build -o ssgo

FROM scratch

COPY --from=build /app/ssgo /usr/bin/ssgo

ENTRYPOINT [ "/usr/bin/ssgo" ]