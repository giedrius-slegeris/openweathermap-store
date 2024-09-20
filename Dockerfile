FROM debian:bookworm

RUN apt-get update && apt-get install -y git wget

WORKDIR /tmp
RUN wget https://go.dev/dl/go1.23.1.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go1.23.1.linux-amd64.tar.gz
RUN rm -rf /tmp/go1.23.1.linux-amd64.tar.gz

ENV PATH=$PATH:/usr/local/go/bin

RUN mkdir -p /git
WORKDIR /git
RUN git clone https://github.com/giedrius-slegeris/openweathermap-store.git

WORKDIR /git/openweathermap-store
RUN go mod download
RUN go build .

CMD ["./openweathermap-store"]