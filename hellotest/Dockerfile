FROM alpine:3.8

RUN mkdir /app
WORKDIR /app
ADD hellotest.tar.gz ./
RUN chmod +x hellotest
CMD ["./hellotest", "server"]
