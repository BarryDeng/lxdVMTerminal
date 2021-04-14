# Compile Stage
FROM barrygates/focal-go1.16:1.0 AS build-env

CMD mkdir /vmconsolews
ADD . /vmconsolews
WORKDIR /vmconsolews

RUN export GOROOT=/usr/local/go && \
    export GOPATH=$HOME/go && \
    export PATH=$GOPATH/bin:$GOROOT/bin:$PATH && \
    go build -o /vmconsolews

# Final Stage
FROM debian:buster

EXPOSE 8084

WORKDIR /
COPY --from=build-env /vmconsolews /

CMD ["/vmconsolews"]