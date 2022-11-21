
# Squid Proxy 


Sample squid proxy and Dockerfile demonstrating various config modes.

The Dockerfile and git image compiles squid with `ssl_crtd` enabled which allows for SSL intercept and rewrite.

The corresponding docker image is on dockerhub:

-  [https://hub.docker.com/r/salrashid123/squidproxy/](https://hub.docker.com/r/salrashid123/squidproxy/)

The image has no entrypoint set to allow you to test and run different modes.

To run the image, simply invoke a shell in the container and start squid in the background for the mode you
are interested in:

This image should be used for debugging and testing so i've left a lot of stuff in the dockerfile (eg, gcc, python, etc).  

If you really want to use this in prod, modify the image to and minimize the footprint (use multistage builds, alpine, `distroless` images, etc)


```bash
docker run  -p 3128:3128 -ti docker.io/salrashid123/squidproxy /bin/bash
```

>> please note that the root CA's have been updated (on `1/9/22`.  You can find the docker image with the original certs as `salrashid123/squidproxy:1` (or you can regenerate your own image from a prior commit))

The CA's provided currently are chained (`root-ca.crt` -> `tls-ca.crt` -> `server_crt.pem`. With the combined root and subordinate as `tls-ca-chain.pem`)


#### changelog

* 11/20/22: Upgrade to debian-11, `squid-5.7`: 
  * `docker.io/salrashid123/squidproxy@sha256:latest`
  * `docker.io/salrashid123/squidproxy@sha256:`

* 1/10/22:  Upgrade built in CA, `squid-3.5.27`:  
  * `docker.io/salrashid123/squidproxy@sha256:b46d3648443d675bb3ac020248495d5d7af1b7f3b683c3068e45c0f040aa5d9c`


Also see
- [Squid proxy cluster with ssl_bump on Google Cloud](https://github.com/salrashid123/squid_ssl_bump_gcp)

### FORWARD

Explicit forward proxy mode intercepts HTTP traffic and uses CONNECT for https.

Launch:

```
$ /apps/squid/sbin/squid -NsY -f /apps/squid.conf.forward &
```

then in a new window run both http and https calls:

```bash
curl -v -x localhost:3128 -L http://httpbin.org/get
curl -v -x localhost:3128 -L https://httpbin.org/get
```

you should see a GET and CONNECT logs within the container

```
$ cat /apps/squid/var/logs/access.log
1668952181.370    112 192.168.9.1 TCP_MISS/301 1560 GET http://www.bbc.com/ - HIER_DIRECT/151.101.0.81 text/html
1668952181.517    146 192.168.9.1 TCP_TUNNEL/200 237935 CONNECT www.bbc.com:443 - HIER_DIRECT/151.101.0.81 -
```

You can also setup allow/deny rules for the domain:
- see [squid.conf.allow_domains](squid.conf.allow_domains)


If you want to use `https_port`, use `squid.conf.https_port`.  For `https_port` see [curl options](https://daniel.haxx.se/blog/2016/11/26/https-proxy-with-curl/) like this:

```bash
curl -v --proxy-cacert tls-ca.crt  --resolve  squid.yourdomain.com:3128:127.0.0.1  -x https://squid.yourdomain.com:3128  https://httpbin.org/get
```


### HTTPS INTERCEPT


In this mode, an HTTPS connection actually terminates the SSL connection _on the proxy_, then proceeds to 
download the certificate for the server you intended to visit.   The proxy server then issues a new certificate with the 
same specifications of the site you wanted to visit and sends that down.

Essentially, the squid proxy is acting as man-in-the-middle.   Ofcourse, you client needs to trust the certificate for the proxy
but if not, you will see a certificate warning.

- [http://www.squid-cache.org/Doc/config/ssl_bump/](http://www.squid-cache.org/Doc/config/ssl_bump/)

Here is the relevant squid conf setting to allow this:

squid.conf.intercept:
```
# Squid normally listens to port 3128
visible_hostname squid.yourdomain.com

http_port 3128 ssl-bump generate-host-certificates=on cert=/apps/tls-ca.crt key=/apps/tls-ca.key

always_direct allow all

acl excluded_sites ssl::server_name .wellsfargo.com
ssl_bump splice excluded_sites
ssl_bump bump all

sslproxy_cert_error deny all
sslcrtd_program /apps/squid/libexec/security_file_certgen -s /apps/squid/var/lib/ssl_db -M 4MB sslcrtd_children 8 startup=1 idle=1
```

The configuration above will insepct all SSL traffic but only _splice_ traffic to wellsfargo.com to view its intended SNI (`server_name`).  You can use the splice capability to apply ACL rules against without inspecting.

- [SslPeekAndSplice](https://wiki.squid-cache.org/Features/SslPeekAndSplice)


Launch
```
$ docker run  -p 3128:3128 -ti docker.io/salrashid123/squidproxy /apps/squid/sbin/squid -NsY -f /apps/squid.conf.intercept
```


then in a new window, try to access a secure site

```bash
$ wget https://raw.githubusercontent.com/salrashid123/squid_proxy/master/tls-ca.crt

$ curl -v --proxy-cacert tls-ca.crt --cacert tls-ca.crt -x localhost:3128  https://www.httpbin.org/get
```

you should see the proxy intercept and recreate httpbin's public certificate:

```
* Server certificate:
*  subject: CN=www.httpbin.org
*  start date: Jan  9 22:05:43 2022 GMT
*  expire date: Jan  9 22:05:43 2032 GMT
*  subjectAltName: host "www.httpbin.org" matched cert's "www.httpbin.org"
*  issuer: C=US; O=Google; OU=Enterprise; CN=Enterprise Subordinate CA       <<<<<<<<<
*  SSL certificate verify ok.
* TLSv1.2 (OUT), TLS header, Supplemental data (23):
> GET /get HTTP/1.1
> Host: www.httpbin.org
> User-Agent: curl/7.85.0
> Accept: */*

```

note the issuer is the proxy's server certificate (`tls-ca.crt`), NOT httpbin's official public cert

Now try to access `www.wellsfargo.com`.  The configuration above simply views the SNI information without snooping on the data

```
$ curl -vvvv --proxy-cacert tls-ca.crt --cacert tls-ca.crt -x localhost:3128  https://www.wellsfargo.com

* Server certificate:
*  subject: jurisdictionC=US; jurisdictionST=Delaware; businessCategory=Private Organization; serialNumber=251212; C=US; ST=California; L=San Francisco; O=Wells Fargo & Company; CN=www.wellsfargo.com
*  start date: Aug  3 00:00:00 2022 GMT
*  expire date: Aug  3 23:59:59 2023 GMT
*  subjectAltName: host "www.wellsfargo.com" matched cert's "www.wellsfargo.com"
*  issuer: C=US; O=DigiCert Inc; CN=DigiCert EV RSA CA G2
*  SSL certificate verify ok.
```

- Also see: [How to Add DNS Filtering to Your NAT Instance with Squid](https://aws.amazon.com/blogs/security/how-to-add-dns-filtering-to-your-nat-instance-with-squid/)

#### Content Adaptation

[content_adaptation/](content_adaptation) allows you to not just intercept SSL traffic, but to actually rewrite the content both ways.

### CACHE

Has cache enabled for HTTP traffic

Launch
```

$ /apps/squid/sbin/squid -NsY -f /apps/squid.conf.cache
```

Run two requests
```
curl -s -x localhost:3128 -L http://www.bbc.com/robots.txt
curl -s -x localhost:3128 -L http://www.bbc.com/robots.txt
```

First request is a TCP_MISS, the second is TCP_MEM_HIT
```
$ cat /apps/squid/var/logs/access.log
1669042557.206     75 192.168.9.1 TCP_MISS/200 20927 GET http://www.bbc.com/robots.txt - HIER_DIRECT/151.101.0.81 text/plain
1669042569.313      0 192.168.9.1 TCP_MEM_HIT/200 20935 GET http://www.bbc.com/robots.txt - HIER_NONE/- text/plain
```

### Basic Auth

Enables squid proxy in default mode but requires a username password for the proxy

 - user: user1
 - password:user1


Launch:

```
$ /apps/squid/sbin/squid -NsY -f /apps/squid.conf.basicauth &
```

```
$ curl -v -x localhost:3128 --proxy-user user1:user1 -L http://httpbin.org/get
```

THe specific config for this mode:

squid.conf.basicaith

```
#user1:user1
#/apps/squid/squid_passwd:  user1:aje5nXwboMxWY
auth_param basic program /apps/squid/libexec/basic_ncsa_auth /apps/squid_passwd
acl authenticated proxy_auth REQUIRED
http_access allow authenticated
http_access deny all
```

Logs would show

```
1669042602.565     37 192.168.9.1 TCP_MISS/200 606 GET http://httpbin.org/get user1 HIER_DIRECT/34.203.186.29 application/json
```


### Dockerfile
```dockerfile
FROM debian:11 AS build
RUN apt-get -y update

RUN apt-get install -y curl supervisor git openssl  build-essential libssl-dev wget vim curl git python3 python3-pip procps
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

RUN chown -R nobody:nogroup /apps/
RUN mkdir -p  /apps/squid/var/lib/
RUN /apps/squid/libexec/security_file_certgen -c -s /apps/squid/var/lib/ssl_db -M 4MB
RUN /apps/squid/sbin/squid -N -f /apps/squid.conf.cache -z
RUN chown -R nobody:nogroup /apps/
RUN chgrp -R 0 /apps && chmod -R g=u /apps
RUN ln -s /usr/bin/python3 /usr/bin/python

EXPOSE 3128
#CMD ["/usr/bin/supervisord"]

```

like i said, its a bit of a large, bloated image.

### Generating new CA

THis repo and image comes with a built-in CA (`root-ca.crt` is the true parent CA that signed a subordinate ca `tls-ca.crt` (yes, i know, its confusing but i used that subca with that name)).  You are free to generate and volume mount your own CA.

- [https://github.com/salrashid123/ca_scratchpad](https://github.com/salrashid123/ca_scratchpad)
