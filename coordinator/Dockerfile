FROM alpine:3.8

RUN mkdir /app
WORKDIR /app
ADD coordinator.tar.gz ./
RUN chmod +x coordinator
CMD ["./coordinator", "daemon"]
