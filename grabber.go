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
    "os"
    "io/ioutil"
    "strings"
    "strconv"
)

const (

    // cgroup_path = "/sys/fs/cgroups/" could be replaced by below
    // To avoid mixing host's data to container(grabber)'s data, we will mount host data to /tmp
    pid_cgroup_path = "/tmp/proc/{pid}/cgroup" //comes out the full path of CPU and RAM
    net_metrics_path = "/tmp/proc/{pid}/net/dev"
    sys_fs_path = "/tmp/sys/fs/cgroup/"
    output_path = "/output/"
    iterates = 10
)

type Grabber struct {
    // The process ID of the container
    pid string
    // The postfix name of output file
    out string
    // The grabbing frequency in millisecond
    ms int
}

func getCgroupMetricPath(cgroup_path string, keyword string) string {

    file_content, err := ioutil.ReadFile(cgroup_path)

    if err != nil {
        fmt.Printf("Error: %v\n", err)
    } else {
        for _, path := range strings.Split(string(file_content), "\n") {
            if strings.Contains(path, keyword) {
                return strings.Split(path, ":")[2]
            }
        }
    }
    return ""

}

func (g Grabber) getCpuData() {

    cpu_dir := "cpu,cpuacct"
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

    f, err := os.Create(output_path + g.out + "_" +fmt.Sprint(g.ms) + "ms_cpu")
    if err != nil {
        log.Fatal(err)
    }

    defer f.Close()

    for i:=0; i<iterates; i++ {
        proc_nums := strings.Fields(outputs[i])
        total_cpu_time := 0
        // add the per cpu seconds
        for _, proc_num := range proc_nums {
            n, _ := strconv.Atoi(proc_num)
            total_cpu_time += n
        }

        f.WriteString(fmt.Sprint(total_cpu_time)+"\n")
    }

    return
}

func findEth0Index(data string) int {
//    data_list = strings.Split(data, "\n")
    for i, d := range strings.Split(data, "\n") {
        if strings.Contains(d, "eth0") {
            return i
        }
    }
    return -1
}

func (g Grabber) getNetworkData() {

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

    eth0_idx := findEth0Index(outputs[0])

    if eth0_idx < 0 {
        log.Print("No eth0's info.")
        return
    }

    for i := 0; i < iterates; i++ {
        metrics := strings.Fields(strings.Split(outputs[i], "\n")[eth0_idx])
        fmt.Printf("%s %s %s %s \n", metrics[1], metrics[2], metrics[9], metrics[10])
    }
}


func main () {

    var metric_type string
    var pid, output_name string
    var interval_ms int


    flag.StringVar(&metric_type, "mtype", "cpu", "What metric to get: cpu/ram/net. (default: cpu)")
    flag.StringVar(&pid, "pid", "0", "The process ID of the container")
    flag.IntVar(&interval_ms, "freq", 5, "The scraping time of metrics collection in millisecond. (default: 5)")
    flag.StringVar(&output_name, "out", "test", "Output name of the metrics")
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
        case "ram" :
            log.Print("Starting to get RAM data")
        case "net" :
            log.Print("Starting to get network data")
            grabber.getNetworkData()
        default:
            log.Fatal("metric_type is not in the handling list")
    }
    //output numbers by type

}
