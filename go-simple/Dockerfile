FROM golang:1.22 AS build

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /main .

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /code
COPY --from=build /main ./main
COPY data.json hello.html ./
EXPOSE 8000
CMD ["/code/main"]
