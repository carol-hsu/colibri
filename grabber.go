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
    "strings"
    "io/ioutil"
)

const (

    // cgroup_path = "/sys/fs/cgroups/" could be replaced by below
    pid_cgroup = "/proc/{pid}/cgroup" //comes out the full path of CPU and RAM
    net_metrics_path = "/proc/{pid}/net/dev" 
    iterates = 10
) 

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

func getCpuData(pid string, ms int, out string) {
    
    path := getCgroupMetricPath(strings.Replace(pid_cgroup, "{pid}", pid, 1), "cpu")

    if path == "" {
        log.Fatal("Error: failed to find the path of CPU data\n")
    }

    var outputs [iterates]string
    for i:=0; i<iterates; i++ {
        v, err  := ioutil.ReadFile(path)
        if err != nil {
            //fmt.Printf("Error: %v\n", err)
            //return
            log.Fatal(err)
        }
        outputs[i] = strings.TrimSpace(string(v))
        time.Sleep(time.Duration(ms) * time.Millisecond)
    }

    
    return
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

    //getting numbers by type
    switch metric_type {
        case "cpu" : 
            log.Print("Starting to get CPU data")
            getCpuData(pid, interval_ms, output_name)
        case "ram" :
            log.Print("Starting to get RAM data")
        case "net" :
            log.Print("Starting to get network data")
        default:
            log.Fatal("metric_type is not in the handling list")
    }
    //output numbers by type


    for i:=0; i < 10; i++ {
        fmt.Printf("get data %d from %s\n", i, pid)
        time.Sleep(time.Duration(interval_ms) * time.Millisecond)
    }

}
