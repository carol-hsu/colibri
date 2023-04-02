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
    PidCgroupPath = "/tmp/proc/{pid}/cgroup" //comes out the full path of CPU and RAM
    NetMetricsPath = "/tmp/proc/{pid}/net/dev"
    VirtualFilesystemPath = "/tmp/cgroup/"

    CpuDirectory = "cpu,cpuacct"
    MemDirectory = "memory"

)

func getCgroupMetricPath(cgroupPath string, keyword string) string {

    content, err := ioutil.ReadFile(cgroupPath)

    if err != nil {
        fmt.Printf("Error: %v\n", err)
    } else if len(keyword) == 0 {
    // v2: return the first line, since it is the only line
        return strings.TrimSpace(string(content))[10:]

    } else {
        for _, path := range strings.Split(string(content), "\n") {
            if strings.Contains(path, keyword) {
                return strings.Split(path, ":")[2]
            }
        }
    }
    return ""

}

func getCpuPath(pid string) string {

    path := getCgroupMetricPath(strings.Replace(PidCgroupPath, "{pid}", pid, 1), CpuDirectory)

    if path == "" {
        log.Fatal("Error: (cgroup v1) failed to find the path of CPU data\n")
    }

    return VirtualFilesystemPath + CpuDirectory + path + "/cpuacct.usage"
}


func getCpuPathV2(pid string) string {

    path := getCgroupMetricPath(strings.Replace(PidCgroupPath, "{pid}", pid, 1), "")

    if path == "" {
        log.Fatal("Error: (cgroup v2) failed to find the path of CPU data\n")
    }

    return VirtualFilesystemPath + path + "/cpu.stat"
}

func getMemPath(pid string) (string, string) {

    path := getCgroupMetricPath(strings.Replace(PidCgroupPath, "{pid}", pid, 1), MemDirectory)

    if path == "" {
        log.Fatal("Error: failed to find the path of Memory data\n")
    }

    return VirtualFilesystemPath + MemDirectory + path + "/memory.usage_in_bytes",
           VirtualFilesystemPath + MemDirectory + path + "/memory.stat"

}

func getMemPathV2(pid string) (string, string) {

    path := getCgroupMetricPath(strings.Replace(PidCgroupPath, "{pid}", pid, 1), "")

    if path == "" {
        log.Fatal("Error: failed to find the path of Memory data\n")
    }

    return VirtualFilesystemPath + path + "/memory.current",
           VirtualFilesystemPath + path + "/memory.stat"

}


func getNetPath(pid string) string {

    return strings.Replace(NetMetricsPath, "{pid}", pid, 1)

}
