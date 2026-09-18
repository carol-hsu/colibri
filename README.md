# Colibri: Fine-grained Resource Profiler for MEC workload

This tool helps you to get the metrics of resource utilization of a specific container in finer-granularity, in millisecond scale.
We do so by getting numbers from the statistics on kernel: through reading the virtual files in `/proc` and `/sys/fs/cgroup`.
Colibri supports both cgroup v1 and v2. **Proved run with Ubuntu 24.04 and Kubernetes v1.31.1**.


## Build the image

<details>
<summary>To verify the cgroup version of your server:</summary>
Checking the system's mount point list to find out which version you are using.
For example, this result below shows v2.
```
$ sudo mount -l | grep "cgroup"
cgroup2 on /sys/fs/cgroup type cgroup2 (rw,nosuid,nodev,noexec,relatime)
```
</details>

You can build image with the root Dockerfile at the root directory. 
It is necessary to indicate the version of cgroup of your system by the parameter `CGROUP_VERSION`.

```
// Building with a image name "colibri" and setting to build for cgroup v2 using the specific files.
$ docker build -t colibri --build-arg CGROUP_VERSION=2 .
```

## Run Colibri job container

After building the image, to run this job-like container, please refer to the key points below:

### Parameters

There are eight dynamic input parameters as following:

| Parameters | Description  | Format  | Default value |
| ---------- | ------------ | ------- | ------------- |
| `name`     | A unique name for the standard metrics output of the specific container. This parameter is used to differentiate the containers in a single Pod. Name it whatever you can recognize for the output      | String | — |
| `pid`      | The process ID of the container. You must specify the correct one to get the metrics you want. [Some instructions](#how-to-get-the-process-id-of-the-container) are described below. | Integer | `0` |
| `mtype`    | The type of metric to collect. There are three of them. `all` will collect all three metric types. | `cpu`, `mem`, `net`, or `all`| `cpu`|
| `span`     | The timespan/sampling interval for getting metrics. The unit is milliseconds. | Integer (millisecondes) | `5` |
| `iter`     | The number of iterations for getting metrics. | Integer | `2000` |
| `out`      | The prefix of output files for raw metrics storage, or the API unique ID for storing the analytic results (for K8s integration). Currently, we only support either of them. You must add either the prefix `file:` or `api:` to indicate the type of output you want. | `none`, `file:<prefix_name>`, or `api:<UUID>` | `none` |
| `iface`    | The network interface of the container from which you want to get metrics. Only used when `mtype` is `net` or `all`. | String | `eth0` |
| `pert`     | The percentile of the metrics shown in standard output. | Integer (percentile) | `95` |

#### More details and examples of the `out` parameter

- `none`: There will be no output file of raw metrics. It will only show aggregate values (average and percentile) on container's STDOUT.
- `file:your_log`: Files named `your_log_*` will be generated and put in the mounted output directory.
- `api:default.my-private-registry-866f6fd9b7-48wq7.1234`: 
A UUID for sending analytics numbers to the Colibri API server for storage. 
The value points to a container with process ID `1234`, running in the Pod `my-private-registry-866f6fd9b7-48wq7` 
in the `default` Namespace. If this information is not correct, the Colibri API server will block this process.

#### How to get the process ID of the container

Before running this tool, you will need to know the process ID of the container on your host.

One method is to check the entry command of the container. 
For example, I want to get the metrics of a container running [Prometheus](https://prometheus.io/), 
and I know its entry command contains `prom`.

```
$ ps aux | grep "prom"
nobody    9189  0.6  0.7 2060936 237084 ?      Ssl  May24  10:40 /bin/prometheus --config.file=/prometheus-cfg/prometheus.yml --storage.tsdb.path=/data
myaccount    22950  0.0  0.0  14428  1024 pts/0    S+   17:30   0:00 grep --color=auto prom
```

Then, we can see that process ID `9189` is for the container.

Or, if you are using Docker to run the containers, 
you can use `docker` command to find the process ID efficiently.

```
// add argument with the specific container name or container ID
$ docker top eaf165466871
UID             PID             PPID            C               STIME           TTY             TIME            CMD
root            94928           94905           0               13:32           pts/0           00:00:00        /bin/bash
```

That's it: `94928` in this case; not PPID, which is for the parent process.


### Mounting points

The virtual file system of cgroup is a significant service in Linux kernel. 
To avoid violating the container environment, 
we prevent to overwrite the them on container.
While the mounting points on container is hardcoded in the program, be awared to mount following directory to the exact pathes (on container).

- The process directory for container ID/directory lookup: `/proc` to `/tmp/proc`
- The prefix directory tree for container metrics: by default, mount `/sys/fs/cgroup/kubepods.slice` to `/tmp/cgroup`.
In fact, the subdirectory of a container in cgroupfs is **NOT** in a fixed style in K8s. 
This can be varied by the versions of cgroup and K8s, and the [QoS](https://kubernetes.io/docs/tasks/configure-pod-container/quality-service-pod/) of the container. 
In cgroup v2 with K8s 1.31, 
    - If container is only set resource requests: change mounting point to `/sys/fs/cgroup/kubepods.slice/kubepods-burstable.slice`  
    - If container isn't set any resource tag: use `/sys/fs/cgroup/kubepods.slice/kubepods-besteffort.slice`
- output file directory to `/output/`

### The example command

Based on previous sections, you can run Colibri job with the carefully configured command.

```
$ docker run -v /proc:/tmp/proc -v /sys/fs/cgroup:/tmp/cgroup -v /my-colibri/log/:/output colibri:latest colibri --pid 1234 --mtype net --span 10 --iter 24000 --out yoman --pert 98
```

### Working with Kubernetes

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


