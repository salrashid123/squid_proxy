
# Squid Proxy 


Sample squid proxy and Dockerfile demonstrating various confg modes.

The Dockerfile and git image compiles squid with ssl_crtd enabled which allows for SSL intercept and rewrite.

The corresponding docker image is on dockerhub:

-  [https://hub.docker.com/r/salrashid123/squidproxy/](https://hub.docker.com/r/salrashid123/squidproxy/)

The image has no entrypoint set to allow you to test and run different modes.

To run the image, simply invoke a shell in the container and start squid in the background for the mode you
are interested in:


```
docker run  -p 3128:3128 -ti docker.io/salrashid123/squidproxy /bin/bash
```

### FORWARD

Explicit forward proxy mode intercepts HTTP traffic and uses CONNECT for https.

Launch:

```
$ /apps/squid/sbin/squid -NsY -f /apps/squid.conf.forward &
```

then in a new window run both http and https calls:

```
$ curl -v -k -x localhost:3128 -L http://www.bbc.com.com/

$ curl -v -k -x localhost:3128 -L https://www.bbc.com.com/
```

you should see a GET and CONNECT logs within the container

```
$ cat /apps/squid/var/logs/access.log
1497880363.186    198 172.17.0.1 TCP_MISS/200 190196 GET http://www.bbc.com/ - HIER_DIRECT/151.101.52.81 text/html
1497880363.439   1392 172.17.0.1 TCP_MISS/200 3403 CONNECT www.bbc.com:443 - HIER_DIRECT/151.101.52.81
```

You can also setup allow/deny rules for the domain:
- see [squid.conf.allow_domains](squid.conf.allow_domains)


### HTTPS INTERCEPT


In this mode, an HTTPS connection actually terminates the SSL connection _on the proxy_, then proceeds to 
download the certificate for the server you intended to visit.   The proxy server then issues a new certificate with the 
same specifications of the site you wanted to visit and sends that down.

Essentially, the squid proxy is acting as man-in-the-middle.   Ofcourse, you client needs to trust the certificate for the proxy
but if not, you will see a certificate warning.

Here is the relevant squid conf setting to allow this:

squid.conf.https_proxy:
```
# Squid normally listens to port 3128
visible_hostname squid.yourdomain.com
http_port 3128 ssl-bump generate-host-certificates=on cert=/apps/server_crt.pem key=/apps/server_key.pem  sslflags=DONT_VERIFY_PEER

always_direct allow all  
ssl_bump server-first all  
sslproxy_cert_error deny all  
sslproxy_flags DONT_VERIFY_PEER  
sslcrtd_program /apps/squid/libexec/ssl_crtd -s /apps/squid/var/lib/ssl_db -M 4MB sslcrtd_children 8 startup=1 idle=1  
```


Launch
```
$ /apps/squid/sbin/squid -NsY -f /apps/squid.conf.https_proxy &
```

then in a new window, try to access a secure site
```
$ curl -v -k -x localhost:3128 -L https://www.yahoo.com
```


you should see the proxy intercept and recreate yahoo's public certificate:


```
* Server certificate:
*      subject: C=US; ST=California; L=Sunnyvale; O=Yahoo Inc.; OU=Information Technology; CN=www.yahoo.com
*      start date: 2015-10-31 00:00:00 GMT
*      expire date: 2017-10-30 23:59:59 GMT
*      issuer: C=US; ST=Illinois; L=Chicago; O=Google Inc.; CN=*.test.google.com
```

note the issuer is the proxy's server certificate (server_crt.pem), NOT yahoo's official public cert

```
$ openssl x509 -in server_crt.pem -text -noout
Certificate:
    Data:
        Version: 3 (0x2)
        Serial Number: 3 (0x3)
    Signature Algorithm: sha1WithRSAEncryption
        Issuer: C=AU, ST=Some-State, O=Internet Widgits Pty Ltd, CN=testca
        Validity
            Not Before: Jul 22 06:00:57 2014 GMT
            Not After : Jul 19 06:00:57 2024 GMT
        Subject: C=US, ST=Illinois, L=Chicago, O=Google Inc., CN=*.test.google.com
        Subject Public Key Info:
            Public Key Algorithm: rsaEncryption
                Public-Key: (1024 bit)
```

- Also see: [How to Add DNS Filtering to Your NAT Instance with Squid](https://aws.amazon.com/blogs/security/how-to-add-dns-filtering-to-your-nat-instance-with-squid/)

### CACHE

Has cache enabled for HTTP traffic

Launch
```
$ /apps/squid/sbin/squid -NsY -f /apps/squid.conf.cache &
```


Run two requests
```
$ curl -k -x localhost:3128 -L http://www.bbc.com

$ curl -k -x localhost:3128 -L http://www.bbc.com
```

First request is a TCP_MISS, the second is TCP_MEM_HIT
```
$ cat /apps/squid/var/logs/access.log
1489070394.303    748 172.17.0.1 TCP_MISS/200 207886 GET http://www.bbc.com/ - HIER_DIRECT/151.101.52.81 text/html
1489070395.767      1 172.17.0.1 TCP_MEM_HIT/200 207721 GET http://www.bbc.com/ - HIER_NONE/- text/html
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
$ curl -x localhost:3128 --proxy-user user1:user1 -L http://www.yahoo.com
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


### Dockerfile
```dockerfile
FROM debian
RUN apt-get -y update && apt-get install -y curl supervisor git openssl  build-essential libssl-dev wget
RUN mkdir -p /var/log/supervisor
COPY supervisord.conf /etc/supervisor/conf.d/supervisord.conf
WORKDIR /apps/
RUN wget -O - http://www.squid-cache.org/Versions/v3/3.4/squid-3.4.14.tar.gz | tar zxfv -
RUN cd /apps/squid-3.4.14/ && ./configure --prefix=/apps/squid --enable-icap-client --enable-ssl --with-openssl --enable-ssl-crtd --enable-auth --enable-basic-auth-helpers="NCS
A" && make && make install
ADD . /apps/

RUN chown -R nobody /apps/
RUN mkdir -p  /apps/squid/var/lib/
RUN /apps/squid/libexec/ssl_crtd -c -s /apps/squid/var/lib/ssl_db -M 4MB
RUN chown -R nobody /apps/

EXPOSE 3128
#CMD ["/usr/bin/supervisord"]
```
