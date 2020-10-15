FROM golang:latest AS build
RUN mkdir /app 
ADD . /app/ 
WORKDIR /app 
RUN \
    go get github.com/gorilla/websocket && \
    go get github.com/friendsofgo/graphiql && \
    go get github.com/graphql-go/graphql
RUN go build -o main . 

FROM golang:latest
RUN mkdir /code
COPY --from=build /app/* /code/
CMD ["/code/main"]