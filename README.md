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

You can build image with the root Dockerfile at the root directory.

```
$ docker build -t colibri .
```

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
- `out`: The prefix of output files for raw metircs storage; or API unique ID for storing the analytic results.
Currently we only support either of them. Must added an another prefix "file:" or "api:" to indicate what kind of special output
you target for.

  By default it is `none`, there will be no output of raw metrics. 

  If the value is assigned to `file:/tmp/colibri/test`, there will come out files named `test_*` and be put at `/tmp/colibri`;

  If the value is assigned to `api:default.my-private-registry-866f6fd9b7-48wq7.1234`, 
it is a uuid for sending analytics numbers to Colibri API server for storage.
The value points to a container with process ID `1234`, 
it is running in the Pod "my-private-registry-866f6fd9b7-48wq7" in "default" Namespace.
If these information is not correct, Colibri API server will block this process.

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

#### Run a standalone job
Please refer to the file `./k8s/colibri.yml` and `./k8s/run_colibri.sh`.

`run_colibri.sh` is a helper script which gives some directions for how to work with the standalone Colibri job:
1. Run your application (marked as `$APP_YAML`).

2. Get the process ID of your application's container 
(fetched by `$CMD_KEYWORD`, and accessed with `$USER` and `$HOSTNAME`, where is running the application).

3. Add the process ID to Colibri K8s YAML.

4. Run Colibri Job, after it is finished, check the metrics querying results.

#### Work with Colibri API server
You can check `./k8s/colibri-api-callback.yml`. 
We will run a job with proper permission attached to it.
The other job configurations are similar to the standalone version. Just be careful of the flag `--out`.


