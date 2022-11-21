# Content Adaptation with ssl_bump

The dockerfile and squid proxy demonstrates two systems:

- [ssl_bump](https://wiki.squid-cache.org/Features/SslPeekAndSplice)
- [content adaptation](https://wiki.squid-cache.org/SquidFaq/ContentAdaptation)

https://github.com/netom/pyicap/pull/42

Unlike the other squid proxies, in this repo, this example runs supervisord to start both the squid proxy and content adaptation server.

The specific content adaptation server is takend from [pyicap](https://github.com/netom/pyicap).

To run, simply build the docker file here and invoke it as



All traffic to the proxy is intercepted but only the following contained in [filter_list.txt](filter_list.txt)"

```
https://www.yahoo.com/robots.txt
http://www.bbc.com/robots.txt
https://cloud.google.com/kubernetes-engine/docs/tutorials/istio-on-gke
https://cloud.google.com/kubernetes-engine/release-notes
```

gets rejected.  

as a demonstration,

build and run:

```bash
## build; note the docker image here is really bloated; if you really want to use this, try to minimize the image size; multistage and distroless images, etc
docker build -t sslproxy .

docker run -ti -p 3128:3128 sslproxy
```


then in a new window, first get the get the cert we used:


```
wget https://raw.githubusercontent.com/salrashid123/squid_proxy/master/content_adaptation/tls-ca.crt
```

A)
```
with proxy:
curl --cacert tls-ca.crt -o /dev/null -s -x localhost:3128 \
  -w "%{http_code}\n" \
  -L https://cloud.google.com/kubernetes-engine/docs/tutorials/istio-on-gke
403
```

B)
```
without proxy:
curl --cacert tls-ca.crt -o /dev/null -s -w "%{http_code}\n" \
  -L https://cloud.google.com/kubernetes-engine/docs/tutorials/istio-on-gke
200
```

C)
```
with proxy:
curl --cacert tls-ca.crt -o /dev/null -s \
  -x localhost:3128 -w "%{http_code}\n" \
  https://cloud.google.com/kubernetes-engine/
200
```


**Note that even over **https** A) is rejected but C) is not...which means squid knew about the URI within the SSL context**