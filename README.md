# Colibri: Fine-grained Metrics Gleaner

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

## Build the image

## Run Colibri job container

After building the image, to run this job-like container, please refer to following key points:

### Parameters

There are four dynamic input parameters as following:
- `name`: A unique name for standard metrics output of the specific container. 
This parameter is used to differenciate the containers in a single Pod.
- `pid`: The process id of the container, must specifying the correct one so to get the metrics you want.
- `mtype`: The types of metric for collection, `cpu`, `mem`, `net` or `all`, `all` will run all three metric types. By default is `cpu`. 
- `freq`: The frequency of getting numbers. The unit is millisecond. By default is `5`. 
- `iter`: The iterations of getting numbers. By default is `2000`. 
- `out`: The prefix of output files. By default it is `none`, there will be no output of raw metrics. If the value is assigned with `test`,
there will come out files named `test_*`
- `iface`: The network interface of the container which you want to get metrics. Only used when `mtype = net`. By default is `eth0`.
- `pert`: The percentile of the metrics shown in standard output. By default is `95`.

### Mounting points

The virual file system of cgroups are the significant service in Linux Kernel, to avoid violating the container environment, we prevent to overwrite the them on container.
While the mounting points on container is hardcoded in the program, be awared to mount following directory to the exact pathes (on container).

- `/proc` to `/tmp/proc`
- `/sys/fs/cgroup` to `/tmp/cgroup`
- output file directory to `/output/`

### The example command

Based on previous sections, you can run Colibri job with the carefully configured command.

```
$ docker run -v /proc:/test/proc -v /sys/fs/cgroup:/tmp/cgroup -v /my-colibri/log/:/output colibri:latest colibri --pid 1234 --mtype net --freq 10 --iter 24000 --out yoman --pert 98
```

### Running with Kubernetes

We can also run our Colibri job through K8s, for getting the metrics on specific workers.
Please refer to the file `./k8s/colibri-job.yml` and `./k8s/run_colibri_job.sh`.
