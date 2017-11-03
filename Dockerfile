FROM golang:1.9.2-stretch

LABEL maintainer phenomenes

ENV PATH=$PATH:$GOPATH/bin

ENV VARNISH_VERSION=5.1.2-1~stretch

RUN apt-get update && apt-get install -y \
	apt-transport-https \
	libjemalloc1 \
	pkg-config \
	&& echo "deb https://packagecloud.io/varnishcache/varnish5/debian/ stretch main" >> /etc/apt/sources.list.d/varnish.list \
	&& curl -s -L https://packagecloud.io/varnishcache/varnish5/gpgkey | apt-key add - \
	&& apt-get update && apt-get install -y \
	varnish=${VARNISH_VERSION} \
	varnish-dev=${VARNISH_VERSION} \
	&& apt-get clean && rm -rf /var/lib/apt/lists/*

RUN mkdir -p $GOPATH/src/github.com/phenomenes/varnishstatbeat

COPY . $GOPATH/src/github.com/phenomenes/varnishstatbeat

WORKDIR $GOPATH/src/github.com/phenomenes/varnishstatbeat

RUN go build .

COPY default.vcl /etc/varnish/default.vcl
COPY docker-entrypoint.sh /docker-entrypoint.sh

RUN sed -i 's/localhost:9200/elasticsearch:9200/' \
	$GOPATH/src/github.com/phenomenes/varnishstatbeat/varnishstatbeat.yml

EXPOSE 8080

CMD /docker-entrypoint.sh
