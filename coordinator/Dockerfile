FROM alpine:3.8

RUN mkdir /app
WORKDIR /app
ADD coordinator ./
CMD ["./coordinator", "daemon"]
