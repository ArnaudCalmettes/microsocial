# This is a multi-stage Dockerfile and requires >= Docker 17.05
# https://docs.docker.com/engine/userguide/eng-image/multistage-build/
FROM gobuffalo/buffalo:v0.14.10 as builder

RUN mkdir -p $GOPATH/src/github.com/ArnaudCalmettes/microsocial
WORKDIR $GOPATH/src/github.com/ArnaudCalmettes/microsocial
RUN go get github.com/gobuffalo/packr/v2 github.com/swaggo/swag/cmd/swag

ADD . .
RUN swag init -g actions/app.go
RUN go get -v ./...
RUN buffalo build --static -o /bin/app

FROM alpine
RUN apk add --no-cache bash
RUN apk add --no-cache ca-certificates

WORKDIR /bin/

COPY --from=builder /bin/app .

# Uncomment to run the binary in "production" mode:
ENV GO_ENV=production

# Bind the app to 0.0.0.0 so it can be seen from outside the container
ENV ADDR=0.0.0.0

EXPOSE 3000

# Uncomment to run the migrations before running the binary:
CMD /bin/app migrate; /bin/app
