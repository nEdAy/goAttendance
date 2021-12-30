FROM scratch

LABEL maintainer="shengsu15@gmail.com" version="1.0"

ADD build/main .
ADD tls ./tls

EXPOSE 8443

CMD ["./main"]