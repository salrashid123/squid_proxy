
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

Also see
- [Squid proxy cluster with ssl_bump on Google Cloud](https://github.com/salrashid123/squid_ssl_bump_gcp)

### FORWARD

Explicit forward proxy mode intercepts HTTP traffic and uses CONNECT for https.

Launch:

```
$ /apps/squid/sbin/squid -NsY -f /apps/squid.conf.forward &
```

then in a new window run both http and https calls:

```
$ curl -v -x localhost:3128 -L http://www.bbc.com/

$ curl -v -x localhost:3128 -L https://www.bbc.com/
```

you should see a GET and CONNECT logs within the container

```
$ cat /apps/squid/var/logs/access.log
1530946085.554    108 172.17.0.1 TCP_MISS/200 224517 GET http://www.bbc.com/ - HIER_DIRECT/151.101.52.81 text/html
1530946085.556    451 172.17.0.1 TCP_TUNNEL/200 3909 CONNECT www.bbc.com:443 - HIER_DIRECT/151.101.52.81 -
```

You can also setup allow/deny rules for the domain:
- see [squid.conf.allow_domains](squid.conf.allow_domains)


If you want to use ```https_port```, use ```squid.conf.https_port```.  For ```https_port``` see [curl options](https://daniel.haxx.se/blog/2016/11/26/https-proxy-with-curl/) like this:

```curl -v --proxy-cacert CA_crt.pem  -x https://squid.yourdomain.com:3128  https://www.yahoo.com/```
(you will need to add `127.0.0.1 squid.yourdomain.com` to your `/etc/hosts` as an override)


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

http_port 3128 ssl-bump generate-host-certificates=on cert=/apps/CA_crt.pem key=/apps/CA_key.pem

always_direct allow all

acl excluded_sites ssl::server_name .wellsfargo.com
ssl_bump splice excluded_sites
ssl_bump bump all

sslproxy_cert_error deny all
sslcrtd_program /apps/squid/libexec/ssl_crtd -s /apps/squid/var/lib/ssl_db -M 4MB sslcrtd_children 8 startup=1 idle=1

```

The configuration above will insepct all SSL traffic but only _splice_ traffic to wellsfargo.com to view its intended SNI (`server_name`).  You can use the splice capability to apply ACL rules against without inspecting.

- [SslPeekAndSplice](https://wiki.squid-cache.org/Features/SslPeekAndSplice)


Launch
```
$ docker run  -p 3128:3128 -ti docker.io/salrashid123/squidproxy /apps/squid/sbin/squid -NsY -f /apps/squid.conf.intercept
```


then in a new window, try to access a secure site
```
$ wget https://raw.githubusercontent.com/salrashid123/squid_proxy/master/CA_crt.pem

$ curl -v --proxy-cacert CA_crt.pem --cacert CA_crt.pem -x localhost:3128  https://www.httpbin.org
```

you should see the proxy intercept and recreate httpbin's public certificate:

```
* Server certificate:
*  subject: CN=www.httpbin.org
*  start date: Dec  6 13:24:37 2019 GMT
*  expire date: Jul 28 13:24:37 2021 GMT
*  subjectAltName: host "www.httpbin.org" matched cert's "www.httpbin.org"
*  issuer: C=US; ST=California; L=Mountain View; O=Google; OU=Enterprise; CN=MyCA   <<<<<<<<<<<<<<<<<<
*  SSL certificate verify ok.
> GET / HTTP/1.1
> Host: www.httpbin.org
> User-Agent: curl/7.66.0
> Accept: */*

```

note the issuer is the proxy's server certificate (`CA_crt.pem`), NOT httpbin's official public cert

Now try to access `www.wellsfargo.com`.  The configuration above simply views the SNI information without snooping on the data

```
$ curl -vvvv --proxy-cacert CA_crt.pem --cacert CA_crt.pem -x localhost:3128  https://www.wellsfargo.com

* Server certificate:
*  subject: businessCategory=Private Organization; jurisdictionC=US; jurisdictionST=Delaware; serialNumber=251212; C=US; ST=California; L=San Francisco; O=Wells Fargo & Company; OU=DCG-PSG; CN=www.wellsfargo.com
*  start date: Feb  8 00:00:00 2019 GMT
*  expire date: Feb  8 12:00:00 2021 GMT
*  subjectAltName: host "www.wellsfargo.com" matched cert's "www.wellsfargo.com"
*  issuer: C=US; O=DigiCert Inc; CN=DigiCert Global CA G2                            <<<<<<<<<<<<<<<<<<
*  SSL certificate verify ok.
> GET / HTTP/1.1
> Host: www.wellsfargo.com
> User-Agent: curl/7.66.0
> Accept: */*
```

- Also see: [How to Add DNS Filtering to Your NAT Instance with Squid](https://aws.amazon.com/blogs/security/how-to-add-dns-filtering-to-your-nat-instance-with-squid/)

#### Content Adaptation

[content_adaptation/](content_adaptation) allows you to not just intercept SSL traffic, but to actually rewrite the content both ways.

### CACHE

Has cache enabled for HTTP traffic

Launch
```
First init the cache directories

$ /apps/squid/sbin/squid -z -f /apps/squid.conf.cache

then
$ /apps/squid/sbin/squid -NsY -f /apps/squid.conf.cache
```
> note: the step to init the cache directory should be in the dockerfile; i've got a todo: to figure out why the setting in the dockerfile itself to init doens't work.


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


### Generating new CA

THis repo and image comes with a built-in CA.  You are free to generate and volume mount your own CA.

To generate your CA, first grab an openssl.cnf file...eg
create openssl.cnf in a folder ('myCA') from here:
 -  https://gist.github.com/salrashid123/dbb032d4538c9df5d5ceed3f3e160ed4

then
```
mkdir myCA
cd myCA
mkdir new_certs
touch index.txt
echo 00 > serial
```

generate the CA:

```
openssl genrsa -out CA_key.pem 2048
openssl req -x509 -days 600 -new -nodes -key CA_key.pem -out CA_crt.pem -extensions v3_ca -config openssl.cnf    -subj "/C=US/ST=California/L=Mountain View/O=Google/OU=Enterprise/CN=MyCA"
```

> note the -extension v3_ca i used and look for that in ```openssl.cnf```

verify
```
openssl x509 -in CA_crt.pem -text -noout
```

```
        X509v3 extensions:
            X509v3 Basic Constraints: 
                CA:TRUE
            X509v3 Key Usage: 
                Digital Signature, Non Repudiation, Key Encipherment, Data Encipherment, Key Agreement, Certificate Sign, CRL Sign
```

---

the ssl_bump mode genrates server certs on the fly for you....but just to complete the steps, if you ever wanted
to generate generate a server cert by hand:

```
openssl genrsa -out server.key 2048
openssl req -config openssl.cnf -days 400 -out server.csr -key server.key -new -sha256  -extensions v3_req  -subj "/C=US/ST=California/L=Mountain View/O=Google/OU=Enterprise/CN=squid.yourdomain.com"
openssl ca -config openssl.cnf -days 400 -notext  -in server.csr   -out server.crt
```

> note the -extension v3_req i used and look for that in ```openssl.cnf```

verify
```
openssl x509 -in server.crt -text -noout
```

```
    Signature Algorithm: sha256WithRSAEncryption
        Issuer: C = US, ST = California, L = Mountain View, O = Google, OU = Enterprise, CN = MyCA
        Validity
            Not Before: May 22 14:27:15 2018 GMT
            Not After : Jun 26 14:27:15 2019 GMT
        Subject: C = US, ST = California, O = Google, OU = Enterprise, CN = squid.yourdomain.com
        X509v3 extensions:
            Netscape Comment: 
                OpenSSL Generated Certificate
            X509v3 Key Usage: 
                Digital Signature, Non Repudiation, Key Encipherment
```


