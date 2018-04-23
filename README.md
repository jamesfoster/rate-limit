# rate-limit

rate-limit is a simple command line tool for echoing stdin to stdout but with a delay between each line.

## usage

```
rate-limit [--port <port>] <rate>
```

* `rate`: the number of lines to echo per second.
* `port`: *Optional*: if specified, listens on the given port for changes to the rate.

## Modifying the rate at runtime.

If you've specified a port then you can perform a HTTP POST request to modify the rate at runtime. For example, the below command will change the rate to 100 lines per second.

```
curl localhost:3456 -d '100'
```

If you want to allow remote access, it is recommended to host it behind a reverse proxy like Apache or nginx.