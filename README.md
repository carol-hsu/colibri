# Finer-grained Metrics Grabber

This tool helps you to get the metrics of resource utilization of a specific container in finer-granularity, in millisecond scale.
We do so by getting numbers from the statistics on kernel: through reading the virtual files in `/proc` and `/sys/fs/cgroup`.

Before running this tool, you will need to know the process id of the container on your host.

One method is refering the entry command of the container. 
For example, I want to get the metrics of the container running Prometheus, and I know its entry command including `prom`.

```
$ ps aux | grep "prom"
nobody    9189  0.6  0.7 2060936 237084 ?      Ssl  May24  10:40 /bin/prometheus --config.file=/prometheus-cfg/prometheus.yml --storage.tsdb.path=/data
myaccount    22950  0.0  0.0  14428  1024 pts/0    S+   17:30   0:00 grep --color=auto prom
```

Then, we can get the process id `9189` is for the container.

## Build docker image

## Run container

After building the image, run this job-like container 

### Mounting points

- `/proc`
- `/sys`
- output file directory
