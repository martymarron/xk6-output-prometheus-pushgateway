xk6-output-prometheus-pushgateway
===
This is a k6 extension for publishing test-run metrics to Prometheus via [Pushgateway](https://prometheus.io/docs/instrumenting/pushing/).\
This extension is fully inspired by [xk6-output-prometheus-remote](https://github.com/grafana/xk6-output-prometheus-remote).\
There might be a circumstance not to enable the "[Remote Write](https://prometheus.io/docs/practices/remote_write/)" feature on your Prometheus instance. In that case, the [Pushgateway](https://prometheus.io/docs/instrumenting/pushing/) and this extension are possibly be an alternative solution.


## Usage
```sh
% xk6 build --with github.com/martymarron/xk6-output-prometheus-pushgateway@latest
% K6_PUSHGATEWAY_URL=http://localhost:9091 \
K6_JOB_NAME=k6_load_testing \
./k6 run \
./script.js \
-o output-prometheus-pushgateway
```

# Prometheus Labels

It is possible to add Prometheus labels with external JSON dictionary in `js` script

```
export const options = {
  ext: {
    "pushgateway": {
      app: "myapp",
      env: "myenv",
    }
  }
};
```

or with environment variables

```
K6_LABEL_APP=myapp K6_LABEL_ENV=myenv k6 run ...
```

It is possible to check Prometheus labels with debug output. It looks like
```
DEBU[0000] Pushgateway labels from JSON options.ext.pushgateway dictionary map[app:MYAPP env:MYENV url:MYURL]
DEBU[0000] Pushgateway labels map[app:MYAPP env:MYENV url:MYURL]
```

The 1st line is Prometheus labels from `options.ext.pushgateway` dictionary.

The 2nd line is the final set of labels after merging `options.ext.pushgateway` dictionary and environment variables.
