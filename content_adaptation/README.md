# Content Adaptation with ssl_bump

The dockerfile and squid proxy demonstrates two systems:

- [ssl_bump](https://wiki.squid-cache.org/Features/SslPeekAndSplice)
- [content adaptation](https://wiki.squid-cache.org/SquidFaq/ContentAdaptation)


Unlike the other squid proxies, in this repo, this example runs supervisord to start both the squid proxy and content adaptation server.

The specific content adaptation server is takend from [pyicap](https://github.com/netom/pyicap).

To run, simply build the docker file here and invoke it as



All traffic to the proxy is intercepted but only the following contained in [filter_list.txt](filter_list.txt)"

```
https://www.yahoo.com/robots.txt
http://www.bbc.com/robots.txt
https://cloud.google.com/kubernetes-engine/docs/tutorials/istio-on-gke
https://cloud.google.com/kubernetes-engine/release-notes
www.cnn.com:443
```

gets rejected.  

as a demonstration,

build and run:

```
docker build -t sslproxy .

docker run -ti -p 3128:3128 sslproxy
```

then in a new window, first get the get the cert we used:

```
wget https://raw.githubusercontent.com/salrashid123/squid_proxy/master/content_adaptation/tls-ca.crt
```
A)
```
curl --cacert tls-ca.crt -x localhost:3128 \
   https://www.cnn.com/
   
curl: (56) Received HTTP code 403 from proxy after CONNECT
```

B)
```
with proxy:
curl --cacert tls-ca.crt -o /dev/null -s -x localhost:3128 \
  -w "%{http_code}\n" \
  -L https://cloud.google.com/kubernetes-engine/docs/tutorials/istio-on-gke
403
```

C)
```
without proxy:
curl --cacert tls-ca.crt -o /dev/null -s -w "%{http_code}\n" \
  -L https://cloud.google.com/kubernetes-engine/docs/tutorials/istio-on-gke
200
```

D)
```
with proxy:
curl --cacert tls-ca.crt -o /dev/null -s \
  -x localhost:3128 -w "%{http_code}\n" \
  -L https://cloud.google.com/kubernetes-engine/
200
```


**Note that even over **https** B) is rejected but D) is not...which means squid knew about the URI within the SSL context**