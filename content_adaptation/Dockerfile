FROM debian:11 AS build
RUN apt-get -y update
RUN apt-get install -y curl supervisor git openssl  build-essential libssl-dev wget python3 python3-pip vim procps
RUN mkdir -p /var/log/supervisor
COPY supervisord.conf /etc/supervisor/conf.d/supervisord.conf
WORKDIR /apps/
RUN wget -O - http://www.squid-cache.org/Versions/v5/squid-5.7.tar.gz | tar zxfv - \
    && CPU=$(( `nproc --all`-1 )) \
    && cd /apps/squid-5.7/ \
    && ./configure --prefix=/apps/squid --enable-icap-client --enable-ssl --with-openssl --enable-ssl-crtd --enable-auth --enable-basic-auth-helpers="NCSA" \
    && make -j$CPU \
    && make install \
    && cd /apps \
    && rm -rf /apps/squid-5.7
ADD . /apps/

RUN chown -R nobody /apps/
RUN mkdir -p  /apps/squid/var/lib/
RUN /apps/squid/libexec/security_file_certgen -c -s /apps/squid/var/lib/ssl_db -M 1MB

RUN wget https://golang.org/dl/go1.19.linux-amd64.tar.gz \
    && tar -zxvf go1.19.linux-amd64.tar.gz -C /usr/local/ \
    && rm go1.19.linux-amd64.tar.gz \
    && /usr/local/go/bin/go mod tidy \
    && /usr/local/go/bin/go build /apps/icap.go 
RUN chown -R nobody /apps/

EXPOSE 3128
CMD ["/usr/bin/supervisord"]

