// Copyright 2022 Carol Hsu
// 
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// 
//     http://www.apache.org/licenses/LICENSE-2.0
// 
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.


package main

import (
    "time"
    "flag"
    "fmt"
    "log"
    "io/ioutil"
    "strings"
)

const (

    // cgroup_path = "/sys/fs/cgroups/" could be replaced by below
    // To avoid mixing host's data to container(grabber)'s data, we will mount host data to /tmp
    pid_cgroup_path = "/tmp/proc/{pid}/cgroup" //comes out the full path of CPU and RAM
    net_metrics_path = "/tmp/proc/{pid}/net/dev"
    sys_fs_path = "/tmp/cgroup/"
    output_path = "/output/"
    iterates = 24000

    cpu_dir = "cpu,cpuacct"
    mem_dir = "memory"
)

type Grabber struct {
    // The process ID of the container
    pid string
    // The postfix name of output file
    out string
    // The grabbing frequency in millisecond
    ms int
}

func (g Grabber) getCpuData() {

    container_path := getCgroupMetricPath(strings.Replace(pid_cgroup_path, "{pid}", g.pid, 1), cpu_dir)

    if container_path == "" {
        log.Fatal("Error: failed to find the path of CPU data\n")
    }

    cpu_data_fullpath := sys_fs_path + cpu_dir + container_path + "/cpuacct.usage_percpu"

    var outputs [iterates]string
    for i:=0; i<iterates; i++ {
        v, err  := ioutil.ReadFile(cpu_data_fullpath)
        if err != nil {
            log.Fatal(err)
        }
        outputs[i] = strings.TrimSpace(string(v))
        time.Sleep(time.Duration(g.ms) * time.Millisecond)
    }

    f := createOutputFile(output_path + g.out + "_" +fmt.Sprint(g.ms) + "ms_cpu")
    defer f.Close()

    for i:=0; i<iterates; i++ {
        proc_nums := strings.Fields(outputs[i])
        total_cpu_time := 0
        // add the per cpu seconds
        for _, proc_num := range proc_nums {
            total_cpu_time += stringToInt(proc_num)
        }

        f.WriteString(fmt.Sprint(total_cpu_time)+"\n")
    }

    return
}

func (g Grabber) getMemoryData() {

    container_path := getCgroupMetricPath(strings.Replace(pid_cgroup_path, "{pid}", g.pid, 1), mem_dir)

    if container_path == "" {
        log.Fatal("Error: failed to find the path of Memory data\n")
    }

    usage_file := sys_fs_path + mem_dir + container_path + "/memory.usage_in_bytes"
    stats_file := sys_fs_path + mem_dir + container_path + "/memory.stat"

    var usage_outputs [iterates]string
    var stats_outputs [iterates]string

    for i:=0; i<iterates; i++ {

        usage, err  := ioutil.ReadFile(usage_file)
        if err != nil {
            log.Fatal(err)
        }
        usage_outputs[i] = strings.TrimSpace(string(usage))

        stats, err  := ioutil.ReadFile(stats_file)
        if err != nil {
            log.Fatal(err)
        }
        stats_outputs[i] = string(stats)

        time.Sleep(time.Duration(g.ms) * time.Millisecond)
    }

    f := createOutputFile(output_path + g.out + "_" +fmt.Sprint(g.ms) + "ms_mem")
    defer f.Close()

    inactive_file_idx := findIndex(stats_outputs[0], "total_inactive_file")

    for i:=0; i<iterates; i++ {
        v := stringToInt(usage_outputs[i]) - stringToInt(strings.Fields(strings.Split(stats_outputs[i], "\n")[inactive_file_idx])[1])
        f.WriteString(fmt.Sprint(v)+"\n")
    }

    return
}

func (g Grabber) getNetworkData(iface string) {

    var outputs [iterates]string
    path := strings.Replace(net_metrics_path, "{pid}", g.pid, 1)

    for i := 0; i < iterates; i++ {
        net_stat, err := ioutil.ReadFile(path)

        if err != nil {
            log.Fatal(err)
        }
        outputs[i] = string(net_stat)
        time.Sleep(time.Duration(g.ms) * time.Millisecond)
    }

    eth0_idx := findIndex(outputs[0], iface)

    if eth0_idx < 0 {
        log.Printf("No info for %s\n", iface)
        return
    }

    recv_bytes_file := createOutputFile(output_path + g.out + "_" +fmt.Sprint(g.ms) + "ms_recv_bytes")
    recv_pkt_file := createOutputFile(output_path + g.out + "_" +fmt.Sprint(g.ms) + "ms_recv_pkt")
    send_bytes_file := createOutputFile(output_path + g.out + "_" +fmt.Sprint(g.ms) + "ms_send_bytes")
    send_pkt_file := createOutputFile(output_path + g.out + "_" +fmt.Sprint(g.ms) + "ms_send_pkt")

    defer recv_bytes_file.Close()
    defer recv_pkt_file.Close()
    defer send_bytes_file.Close()
    defer send_pkt_file.Close()

    // create output files
    for i := 0; i < iterates; i++ {
        metrics := strings.Fields(strings.Split(outputs[i], "\n")[eth0_idx])
        recv_bytes_file.WriteString(metrics[1]+"\n")
        recv_pkt_file.WriteString(metrics[2]+"\n")
        send_bytes_file.WriteString(metrics[9]+"\n")
        send_pkt_file.WriteString(metrics[10]+"\n")
    }
    return
}


func main () {

    var metric_type string
    var pid, output_name, net_iface string
    var interval_ms int


    flag.StringVar(&metric_type, "mtype", "cpu", "What metric to get: cpu/mem/net. (default: cpu)")
    flag.StringVar(&pid, "pid", "0", "The process ID of the container")
    flag.IntVar(&interval_ms, "freq", 5, "The scraping time of metrics collection in millisecond. (default: 5)")
    flag.StringVar(&output_name, "out", "test", "Output name of the metrics")
    flag.StringVar(&net_iface, "iface", "eth0", "The name of network interface of the container. Only used for grabbing network metrics. (default: eth0)")
    flag.Parse()

    if interval_ms <= 0 {
        log.Print("Monitoring process cannot be processed with interval_ms less and equal 0.")
        return
    }

    grabber := Grabber{pid, output_name, interval_ms}
    //getting numbers by type
    switch metric_type {
        case "cpu" :
            log.Print("Starting to get CPU data")
            grabber.getCpuData()
        case "mem" :
            log.Print("Starting to get RAM data")
            grabber.getMemoryData()
        case "net" :
            log.Print("Starting to get network data")
            grabber.getNetworkData(net_iface)
        default:
            log.Fatal("metric_type is not in the handling list")
    }

}
