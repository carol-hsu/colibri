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
    "fmt"
    "strings"
    "log"
    "io/ioutil"
)

const (

    // cgroup_path = "/sys/fs/cgroups/" could be replaced by below
    // To avoid mixing host's data to container(scraper)'s data, we will mount host data to /tmp
    pid_cgroup_path = "/tmp/proc/{pid}/cgroup" //comes out the full path of CPU and RAM
    net_metrics_path = "/tmp/proc/{pid}/net/dev"
    sys_fs_path = "/tmp/cgroup/"

    cpu_dir = "cpu,cpuacct"
    mem_dir = "memory"

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

func getCpuPath(pid string) string {

    container_path := getCgroupMetricPath(strings.Replace(pid_cgroup_path, "{pid}", pid, 1), cpu_dir)

    if container_path == "" {
        log.Fatal("Error: failed to find the path of CPU data\n")
    }

    return sys_fs_path + cpu_dir + container_path + "/cpuacct.usage"
}

func getMemPath(pid string) (string, string) {

    container_path := getCgroupMetricPath(strings.Replace(pid_cgroup_path, "{pid}", pid, 1), mem_dir)

    if container_path == "" {
        log.Fatal("Error: failed to find the path of Memory data\n")
    }

    return sys_fs_path + mem_dir + container_path + "/memory.usage_in_bytes",
           sys_fs_path + mem_dir + container_path + "/memory.stat"

}

func getNetPath(pid string) string {

    return strings.Replace(net_metrics_path, "{pid}", pid, 1)

}
