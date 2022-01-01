FROM alpine

LABEL maintainer="shengsu15@gmail.com" version="1.0"

ADD build/main .
ADD tls ./tls
ADD app.properties .

RUN apk add --no-cache ca-certificates && \
    update-ca-certificates

EXPOSE 8443
CMD ["./main"]