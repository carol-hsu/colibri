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

const output_path = "/output/"

func (s Scraper) getCpuData() []float64 {

    cpu_data_fullpath := getCpuPath(s.pid)

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

    return countRate(outputs, s.ms, s.pert)
}

func (s Scraper) getMemoryData() []float64 {

    usage_file, stats_file := getMemPath(s.pid)

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

    return countValue(outputs, s.pert)
}

func (s Scraper) getNetworkData(iface string) []float64 {

    var outputs =[]string{}
    path := getNetPath(s.pid)

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
        log.Fatal("No info for the specified interface")
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


    return append(ig_res, eg_res...)
}

func getCpuValue(path string) string {

    v, err  := ioutil.ReadFile(path)
    if err != nil {
        log.Print("Cannot read usage file of cpu.")
        return ""
    }
    return strings.TrimSpace(string(v))
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

func getInactiveFileIndex(path string) int {

    stats, err  := ioutil.ReadFile(path)
    if err != nil {
        log.Print("Cannot read statistic file of memory.")
        return -1
    }

    return findIndex(string(stats), "total_inactive_file")
}

func getNetworkValue(path string, idx int) (string, string) {

    net_stat, err := ioutil.ReadFile(path)
    if err != nil {
        log.Print("Cannot read statistic file of network.")
        return "", ""
    }
    stats := strings.Fields(strings.Split(string(net_stat), "\n")[idx])

    return stats[1], stats[9]
}

func getIfaceIndex(path string, iface string) int {

    stats, err  := ioutil.ReadFile(path)
    if err != nil {
        log.Print("Cannot read statistic file of network.")
        return -1
    }

    return findIndex(string(stats), iface)
}

func (s Scraper) getAllData(iface string) ([]float64, []float64, []float64) {
    //get path of container
    cpu_path := getCpuPath(s.pid)
    usage_path, stats_path := getMemPath(s.pid)
    net_path := getNetPath(s.pid)

    var cpu_outputs, ig_outputs, eg_outputs []string
    var mem_outputs = []float64{}

    //get index for collecting data from memory statistic file
    mem_idx := getInactiveFileIndex(stats_path)
    net_idx := getIfaceIndex(net_path, iface)

    //start metrics scraping period
    for i:=0; i < s.iter; i++ {
        cpu_v := getCpuValue(cpu_path)
        mem_v := getMemoryValue(usage_path, stats_path, mem_idx)
        ig_bw, eg_bw := getNetworkValue(net_path, net_idx)

        if mem_v < 0 || len(cpu_v) == 0 || len(ig_bw) == 0 {
            log.Print("App stopped earlier, starting to print output")
            break
        }

        cpu_outputs = append(cpu_outputs, cpu_v)
        mem_outputs = append(mem_outputs, mem_v)
        ig_outputs = append(ig_outputs, ig_bw)
        eg_outputs = append(eg_outputs, eg_bw)

        time.Sleep(time.Duration(s.ms) * time.Millisecond)
    }

    //if output_name == none, then don't write out, just print analysis result
    if s.out != "none" {
        file_prefix := output_path + s.out + "_" +fmt.Sprint(s.ms)

        cpu_f := createOutputFile(file_prefix + "ms_cpu")
        defer cpu_f.Close()

        mem_f := createOutputFile(file_prefix + "ms_mem")
        defer mem_f.Close()

        ig_f := createOutputFile(file_prefix + "ms_ig_bytes")
        defer ig_f.Close()

        eg_f := createOutputFile(file_prefix + "ms_eg_bytes")
        defer eg_f.Close()

        for i := 0; i < len(cpu_outputs); i++ {
            cpu_f.WriteString(cpu_outputs[i]+"\n")
            mem_f.WriteString(fmt.Sprintf("%.0f\n", mem_outputs[i]))
            ig_f.WriteString(ig_outputs[i]+"\n")
            eg_f.WriteString(eg_outputs[i]+"\n")
        }
    }

    cpu_res := countRate(cpu_outputs, s.ms, s.pert)
    mem_res := countValue(mem_outputs, s.pert)
    ig_res := countRate(ig_outputs, s.ms, s.pert)
    eg_res := countRate(eg_outputs, s.ms, s.pert)

    return cpu_res, mem_res, append(ig_res, eg_res...)
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
            res := scraper.getCpuData()
            log.Printf("%s -- CPU Avg: %d, %.2f-Percentile: %d\n", name,
                                                                   int(math.Round(res[0]/1000)), percentile,
                                                                   int(math.Round(res[1]/1000)))

        case "mem" :
            log.Print("Starting to get RAM data")
            res := scraper.getMemoryData()
            log.Printf("%s -- RAM Avg: %d, %.2f-Percentile: %d\n", name,
                                                                   int(math.Round(res[0]/1024/1024)), percentile,
                                                                   int(math.Round(res[1]/1024/1024)))

        case "net" :
            log.Print("Starting to get network data")
            res := scraper.getNetworkData(net_iface)
            log.Printf("%s -- Ingress Avg: %s, %.2f-Percentile: %s\n", name,
                                                                       transBandwidthUnit(res[0]), percentile,
                                                                       transBandwidthUnit(res[1]))
            log.Printf("%s -- Egress Avg: %s, %.2f-Percentile: %s\n", name,
                                                                      transBandwidthUnit(res[2]), percentile,
                                                                      transBandwidthUnit(res[3]))

        case "all":
            log.Print("Starting to get all metrics: TEST")
            cpu_res, mem_res, net_res := scraper.getAllData(net_iface)

            log.Printf("%s -- CPU Avg: %d, %.2f-Percentile: %d\n", name,
                                                                   int(math.Round(cpu_res[0]/1000)), percentile,
                                                                   int(math.Round(cpu_res[1]/1000)))
            log.Printf("%s -- RAM Avg: %d, %.2f-Percentile: %d\n", name,
                                                                   int(math.Round(mem_res[0]/1024/1024)), percentile,
                                                                   int(math.Round(mem_res[1]/1024/1024)))
            log.Printf("%s -- NET Ingress Avg: %s, %.2f-Percentile: %s\n", name,
                                                                           transBandwidthUnit(net_res[0]), percentile,
                                                                           transBandwidthUnit(net_res[1]))
            log.Printf("%s -- NET Egress Avg: %s, %.2f-Percentile: %s\n", name,
                                                                          transBandwidthUnit(net_res[2]), percentile,
                                                                          transBandwidthUnit(net_res[3]))

        default:
            log.Fatal("metric_type is not in the handling list")
    }

    log.Print("Colibri is successfully completed !")

}
