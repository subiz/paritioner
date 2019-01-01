FROM alpine:3.8

RUN mkdir /app
WORKDIR /app
ADD hello_test.tar.gz ./
RUN chmod +x hello_test
CMD ["./hello_test", "server"]
