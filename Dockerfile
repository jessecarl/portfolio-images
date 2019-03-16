FROM golang:1.12 as build

WORKDIR /go/src/github.com/jessecarl/portfolio-images
COPY . .
ENV GO111MODULE="on"

RUN go get -d -v ./...
RUN go install -v ./...

FROM gcr.io/distroless/base
COPY --from=build /go/bin/portfolio-images /
CMD ["/portfolio-images"]
