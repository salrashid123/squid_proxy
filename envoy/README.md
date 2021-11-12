### Envoy CONNECT proxy

Example envoy proxy config that acts as a basic `CONNECT` mode 


```bash
envoy  version: c5594b41f48b2e45df83105de84623f2930d23cd/1.15.0/clean-getenvoy-a5345f6-envoy/RELEASE/Bo


./envoy -c basic.yaml -l debug

```

then in another window

```
curl -vv -x localhost:3128 -L https://httpbin.org/get
```