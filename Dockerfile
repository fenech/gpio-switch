FROM arm32v7/golang:alpine as build

WORKDIR /opt/app

COPY go.mod go.sum ./
RUN go mod download

COPY main.go .
RUN CGO_ENABLED=0 go build -a -o gpio-switch .

FROM scratch

COPY --from=build /opt/app/gpio-switch ./
ENTRYPOINT [ "./gpio-switch" ]
