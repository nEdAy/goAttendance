FROM scratch

LABEL maintainer="shengsu15@gmail.com" version="1.0"

ADD build/main .

EXPOSE 9443

CMD ["./main"]