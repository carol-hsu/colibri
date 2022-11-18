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
// See the License for the specific language s.verning permissions and
// limitations under the License.

package main

import (
    "time"
    "flag"
    "fmt"
    "log"
    "math"
    "io/ioutil"
    "strings"
)

const (

    // cgroup_path = "/sys/fs/cgroups/" could be replaced by below
    // To avoid mixing host's data to container(scraper)'s data, we will mount host data to /tmp
    pid_cgroup_path = "/tmp/proc/{pid}/cgroup" //comes out the full path of CPU and RAM
    net_metrics_path = "/tmp/proc/{pid}/net/dev"
    sys_fs_path = "/tmp/cgroup/"
    output_path = "/output/"

    cpu_dir = "cpu,cpuacct"
    mem_dir = "memory"

)

type Scraper struct {
    // The process ID of the container
    pid string
    // The postfix name of output file
    out string
    // The s.abbing frequency in millisecond
    ms int
    // The s.abbing iterates
    iter int
    // The percentile for output
    pert float64
}

func (s Scraper) getCpuData(c chan []float64) {

    container_path := getCgroupMetricPath(strings.Replace(pid_cgroup_path, "{pid}", s.pid, 1), cpu_dir)

    if container_path == "" {
        log.Fatal("Error: failed to find the path of CPU data\n")
    }

    cpu_data_fullpath := sys_fs_path + cpu_dir + container_path + "/cpuacct.usage"

    var outputs = []string{}

    for i:=0; i<s.iter; i++ {
        v, err  := ioutil.ReadFile(cpu_data_fullpath)
        if err != nil {
            if i == 0 {
            // nothing existed in output, then forcefully stop
                log.Fatal(err)
            }else{
                log.Print("App stopped earlier, starting to print output")
                break
            }
        }
        outputs = append(outputs, strings.TrimSpace(string(v)))
        time.Sleep(time.Duration(s.ms) * time.Millisecond)
    }

    log.Print("CPU metrics collection is finished. Start to post-process data ...")
    //if output_name == none, then don't write out, just print analysis result
    if s.out != "none" {
        f := createOutputFile(output_path + s.out + "_" +fmt.Sprint(s.ms) + "ms_cpu")
        defer f.Close()

        for i:=0; i<len(outputs); i++ {
            f.WriteString(outputs[i]+"\n")
        }
    }

    res := countRate(outputs, s.ms, s.pert)

    if c != nil{
        //sending for done
        c <- res
    }else{
        log.Printf("CPU Avg: %d, %.2f-Percentile: %d\n", int(math.Round(res[0]/1000)), s.pert,
                                                         int(math.Round(res[1]/1000)))
    }

    return
}

func (s Scraper) getMemoryData(c chan []float64) {

    container_path := getCgroupMetricPath(strings.Replace(pid_cgroup_path, "{pid}", s.pid, 1), mem_dir)

    if container_path == "" {
        log.Fatal("Error: failed to find the path of Memory data\n")
    }

    usage_file := sys_fs_path + mem_dir + container_path + "/memory.usage_in_bytes"
    stats_file := sys_fs_path + mem_dir + container_path + "/memory.stat"

    var usage_outputs, stats_outputs = []string{}, []string{}

    for i:=0; i < s.iter; i++ {

        usage, err  := ioutil.ReadFile(usage_file)
        if err != nil {
            if i == 0 {
            // nothing existed in output, then forcefully stop
                log.Fatal(err)
            }else{
                log.Print("App stopped earlier, starting to print output")
                break
            }
        }
        usage_outputs = append(usage_outputs, strings.TrimSpace(string(usage)))

        stats, err  := ioutil.ReadFile(stats_file)
        if err != nil {
            if i == 0 {
            // nothing existed in output, then forcefully stop
                log.Fatal(err)
            }else{
                log.Print("App stopped earlier, starting to print output")
                break
            }
        }
        stats_outputs = append(stats_outputs, string(stats))

        time.Sleep(time.Duration(s.ms) * time.Millisecond)
    }

    log.Print("Memory metrics collection is finished. Start to post-process data ...")

    //count usage and inactive file size, and stored in float
    var outputs = make([]float64, len(stats_outputs))

    inactive_file_idx := findIndex(stats_outputs[0], "total_inactive_file")

    for i := 0; i < len(outputs); i++ {
        v := stringToFloat(usage_outputs[i]) - stringToFloat(strings.Fields(strings.Split(stats_outputs[i], "\n")[inactive_file_idx])[1])
        outputs[i] = v
    }

    //if output_name == none, then don't write out, just print analysis result
    if s.out != "none" {
        f := createOutputFile(output_path + s.out + "_" +fmt.Sprint(s.ms) + "ms_mem")
        defer f.Close()


        for i := 0; i < len(outputs); i++ {
            f.WriteString(fmt.Sprintf("%f\n", outputs[i]))
        }
    }

    res := countValue(outputs, s.pert)

    if c != nil{
        //sending for done
        c <- res
    }else{
        // print result
        log.Printf("RAM Avg: %d, %.2f-Percentile: %d\n", int(math.Round(res[0]/1024/1024)), s.pert,
                                                         int(math.Round(res[1]/1024/1024)))
    }
    return
}

func (s Scraper) getNetworkData(iface string, c chan []float64) {

    var outputs =[]string{}
    path := strings.Replace(net_metrics_path, "{pid}", s.pid, 1)

    for i := 0; i < s.iter; i++ {
        net_stat, err := ioutil.ReadFile(path)

        if err != nil {
            if i == 0 {
            // nothing existed in output, then forcefully stop
                log.Fatal(err)
            }else{
                log.Print("App stopped earlier, starting to print output")
                break
            }
        }
        outputs = append(outputs, string(net_stat))
        time.Sleep(time.Duration(s.ms) * time.Millisecond)
    }

    log.Print("Network metrics collection is finished. Start to post-process data ...")

    eth0_idx := findIndex(outputs[0], iface)

    if eth0_idx < 0 {
        log.Printf("No info for %s\n", iface)
        return
    }

    //parse bandwidth value, and store separately
    var output_len = len(outputs)
    var ig_bw = make([]string, output_len)
    var eg_bw = make([]string, output_len)

    for i := 0; i < output_len; i++ {
        metrics := strings.Fields(strings.Split(outputs[i], "\n")[eth0_idx])
        ig_bw[i] = metrics[1]
        eg_bw[i] = metrics[9]
    }

    //if output_name == none, then don't write out, just print analysis result
    if s.out != "none" {
        ig_file := createOutputFile(output_path + s.out + "_" + fmt.Sprint(s.ms) + "ms_ig_bytes")
        eg_file := createOutputFile(output_path + s.out + "_" + fmt.Sprint(s.ms) + "ms_eg_bytes")

        defer ig_file.Close()
        defer eg_file.Close()

        // create output files
        for i := 0; i < output_len; i++ {
            ig_file.WriteString(ig_bw[i]+"\n")
            eg_file.WriteString(eg_bw[i]+"\n")
        }
    }

    ig_res := countRate(ig_bw, s.ms, s.pert)
    eg_res := countRate(eg_bw, s.ms, s.pert)

    if c != nil{
        c <- append(ig_res, eg_res...)
        return
    }else {
        // print result
        log.Printf("Ingress Avg: %s, %.2f-Percentile: %s\n", transBandwidthUnit(ig_res[0]), s.pert,
                                                             transBandwidthUnit(ig_res[1]))
        log.Printf("Egress Avg: %s, %.2f-Percentile: %s\n", transBandwidthUnit(eg_res[0]), s.pert,
                                                            transBandwidthUnit(eg_res[1]))
    }

    return
}

func getMemoryValue(usage_path string, stats_path string, idx int) float64 {

    usage, err  := ioutil.ReadFile(usage_path)
    if err != nil {
        log.Print("Cannot read usage file of memory.")
        return -1
    }

    usage_output := strings.TrimSpace(string(usage))

    stats, err  := ioutil.ReadFile(stats_path)
    if err != nil {
        log.Print("Cannot read statistic file of memory.")
        return -1
    }
    stats_output := string(stats)

    return stringToFloat(usage_output) - stringToFloat(strings.Fields(strings.Split(stats_output, "\n")[idx])[1])
}

func getInactiveFileIndex(stats_path string) int {

    stats, err  := ioutil.ReadFile(stats_path)
    if err != nil {
        log.Print("Cannot read statistic file of memory.")
        return -1
    }

    return findIndex(string(stats), "total_inactive_file")
}

func (s Scraper) getAllData(iface string) {
    //test for collect memory first ..
    container_path := getCgroupMetricPath(strings.Replace(pid_cgroup_path, "{pid}", s.pid, 1), mem_dir)

    if container_path == "" {
        log.Fatal("Error: failed to find the path of Memory data\n")
    }

}

func main () {

    var metric_type, name, pid, output_name, net_iface string
    var interval_ms, iterate_num int
    var percentile float64

    flag.StringVar(&name, "name", "birdy", "The name of this work to indicate for standard output. (default: birdy)")
    flag.StringVar(&metric_type, "mtype", "cpu", "What metric to s.t: cpu/mem/net/all. (default: cpu)")
    flag.StringVar(&pid, "pid", "0", "The process ID of the container")
    flag.IntVar(&interval_ms, "freq", 5, "The scraping interval in millisecond. (default: 5)")
    flag.IntVar(&iterate_num, "iter", 2000, "The scraping numbers. (default: 2000)")
    flag.Float64Var(&percentile, "pert", 95, "The percentile value for analytics. (default: 95)")
    flag.StringVar(&output_name, "out", "none", "Output file for the metrics")
    flag.StringVar(&net_iface, "iface", "eth0", "The name of network interface of the container. Only used for s.abbing network metrics. (default: eth0)")
    flag.Parse()

    if interval_ms <= 0 {
        log.Print("Monitoring process cannot be processed with interval_ms less and equal 0.")
        return
    }
    scraper := Scraper{pid, output_name, interval_ms, iterate_num, percentile}

    //getting numbers by type
    switch metric_type {
        case "cpu" :
            log.Print("Starting to get CPU data")
            scraper.getCpuData(nil)

        case "mem" :
            log.Print("Starting to get RAM data")
            scraper.getMemoryData(nil)

        case "net" :
            log.Print("Starting to get network data")
            scraper.getNetworkData(net_iface, nil)

        case "all":
            log.Print("Starting to get all metrics")
            cpu_c, mem_c, net_c := make(chan []float64), make(chan []float64), make(chan []float64)
            go scraper.getCpuData(cpu_c)
            go scraper.getMemoryData(mem_c)
            go scraper.getNetworkData(net_iface, net_c)
            cpu_out, mem_out, net_out := <-cpu_c, <-mem_c, <-net_c
            log.Printf("%s -- CPU Avg: %d, %.2f-Percentile: %d\n", name,
                                                                int(math.Round(cpu_out[0]/1000)), percentile,
                                                                int(math.Round(cpu_out[1]/1000)))
            log.Printf("%s -- RAM Avg: %d, %.2f-Percentile: %d\n", name,
                                                                int(math.Round(mem_out[0]/1024/1024)), percentile,
                                                                int(math.Round(mem_out[1]/1024/1024)))
            log.Printf("%s -- NET Ingress Avg: %s, %.2f-Percentile: %s\n", name,
                                                                        transBandwidthUnit(net_out[0]), percentile,
                                                                        transBandwidthUnit(net_out[1]))
            log.Printf("%s -- NET Egress Avg: %s, %.2f-Percentile: %s\n", name,
                                                                       transBandwidthUnit(net_out[2]), percentile,
                                                                       transBandwidthUnit(net_out[3]))
        case "test":
            log.Print("Starting to get all metrics")
            scraper.getAllData(net_iface)
        default:
            log.Fatal("metric_type is not in the handling list")
    }

    log.Print("Colibri is successfully completed !")

}
