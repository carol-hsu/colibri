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
    "io/ioutil"
    "strings"
)

type Scraper struct {
    // The process ID of the container
    pid string
    // The postfix name of output file
    out string
    // The metric scraping timespan in millisecond
    ms int
    // The scrapping #iteration
    iter int
    // The percentile for output
    pert float64
}

const output_path = "/output/"

func (s Scraper) getCpuData() []float64 {

    cpu_data_fullpath := getCpuPathV2(s.pid)
    var stats_outputs = []string{}

    for i:=0; i<s.iter; i++ {
        stats, err  := ioutil.ReadFile(cpu_data_fullpath)
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

    log.Print("CPU metrics collection is finished. Start to post-process data ...")

    var outputs = make([]string, len(stats_outputs))

    usage_idx := findIndex(stats_outputs[0], "usage_usec")

    for i := 0; i < len(outputs); i++ {
        outputs[i] = strings.Fields(strings.Split(stats_outputs[i], "\n")[usage_idx])[1]
    }
    //if outputName == none, then don't write out, just print analysis result
    if strings.Contains(s.out, "file:") {
        f := createOutputFile(output_path + s.out[5:] + "_" +fmt.Sprint(s.ms) + "ms_cpu")
        defer f.Close()

        for i:=0; i<len(outputs); i++ {
            f.WriteString(outputs[i]+"\n")
        }
    }

    return countRate(outputs, s.ms, s.pert)
}

func (s Scraper) getMemoryData() []float64 {

    usage_file, stats_file := getMemPathV2(s.pid)

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

    inactive_file_idx := findIndex(stats_outputs[0], "inactive_file")

    for i := 0; i < len(outputs); i++ {
        v := stringToFloat(usage_outputs[i]) - stringToFloat(strings.Fields(strings.Split(stats_outputs[i], "\n")[inactive_file_idx])[1])
        outputs[i] = v
    }

    //if outputName == none, then don't write out, just print analysis result
    if strings.Contains(s.out, "file:") {
        f := createOutputFile(output_path + s.out[5:] + "_" +fmt.Sprint(s.ms) + "ms_mem")
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

    //if outputName == none, then don't write out, just print analysis result
    if strings.Contains(s.out, "file:") {
        ig_file := createOutputFile(output_path + s.out[5:] + "_" + fmt.Sprint(s.ms) + "ms_ig_bytes")
        eg_file := createOutputFile(output_path + s.out[5:] + "_" + fmt.Sprint(s.ms) + "ms_eg_bytes")

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

func getCpuValue(path string, idx int) string {

    stats, err  := ioutil.ReadFile(path)
    if err != nil {
        log.Print("Cannot read statistic file of cpu: ", err)
        return ""
    }

    return strings.Fields(strings.Split(string(stats), "\n")[idx])[1]
}

func getMemoryValue(usage_path string, stats_path string, idx int) float64 {

    usage, err  := ioutil.ReadFile(usage_path)
    if err != nil {
        log.Print("Cannot read usage file of memory: ", err)
        return -1
    }

    usage_output := strings.TrimSpace(string(usage))

    stats, err  := ioutil.ReadFile(stats_path)
    if err != nil {
        log.Print("Cannot read statistic file of memory: ", err)
        return -1
    }

    return stringToFloat(usage_output) - stringToFloat(strings.Fields(strings.Split(string(stats), "\n")[idx])[1])
}

func getUsageIndex(path string) int {

    stats, err  := ioutil.ReadFile(path)
    if err != nil {
        log.Print("Cannot read statistic file of cpu: ", err)
        return -1
    }

    return findIndex(string(stats), "usage_usec")
}


func getInactiveFileIndex(path string) int {

    stats, err  := ioutil.ReadFile(path)
    if err != nil {
        log.Print("Cannot read statistic file of memory: ", err)
        return -1
    }

    return findIndex(string(stats), "inactive_file")
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
    cpu_path := getCpuPathV2(s.pid)
    usage_path, stats_path := getMemPathV2(s.pid)
    net_path := getNetPath(s.pid)

    var cpu_outputs, ig_outputs, eg_outputs []string
    var mem_outputs = []float64{}

    //get index for collecting data from memory statistic file
    cpu_idx := getUsageIndex(cpu_path)
    mem_idx := getInactiveFileIndex(stats_path)
    net_idx := getIfaceIndex(net_path, iface)

    //start metrics scraping period
    for i:=0; i < s.iter; i++ {
        cpu_v := getCpuValue(cpu_path, cpu_idx)
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

    //if outputName == none, then don't write out, just print analysis result
    if strings.Contains(s.out, "file:") {
        file_prefix := output_path + s.out[5:] + "_" +fmt.Sprint(s.ms)

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

    var metricType, name, pid, outputName, netIface string
    var intervalMsec, iterateNum int
    var percentile float64

    flag.StringVar(&name, "name", "birdy", "The name of this work to indicate for standard output. (default: birdy)")
    flag.StringVar(&metricType, "mtype", "cpu", "What metric to s.t: cpu/mem/net/all. (default: cpu)")
    flag.StringVar(&pid, "pid", "0", "The process ID of the container")
    flag.IntVar(&intervalMsec, "span", 5, "The scraping interval/timespan in millisecond. (default: 5)")
    flag.IntVar(&iterateNum, "iter", 2000, "The scraping numbers. (default: 2000)")
    flag.Float64Var(&percentile, "pert", 95, "The percentile value for analytics. (default: 95)")
    flag.StringVar(&outputName, "out", "none", "Output file or API unique ID for storing the metrics")
    flag.StringVar(&netIface, "iface", "eth0", "The name of network interface of the container. Only used for s.abbing network metrics. (default: eth0)")
    flag.Parse()

    log.SetFlags(log.LstdFlags | log.Lmicroseconds)

    if intervalMsec <= 0 {
        log.Print("Monitoring process cannot be processed with intervalMsec less and equal 0.")
        return
    }
    scraper := Scraper{pid, outputName, intervalMsec, iterateNum, percentile}

    //getting numbers by type
    switch metricType {
        case "cpu" :
            log.Print("Starting to get CPU data")
            res := scraper.getCpuData()
            pertRes := transCpuUnitV2(res[1])
            printResult(name, "CPU", transCpuUnitV2(res[0]), pertRes, percentile)

            if scraper.out[:4] == "api:" {
                log.Println("Calling API!")
                sendMetric([]byte(`{ "cpu" : "` + pertRes + `" }`), scraper.out[4:])
            }

        case "mem" :
            log.Print("Starting to get RAM data")
            res := scraper.getMemoryData()
            pertRes := transMemoryUnit(res[1])
            printResult(name, "RAM", transMemoryUnit(res[0]), pertRes, percentile)

            if scraper.out[:4] == "api:" {
                log.Println("Calling API!")
                sendMetric([]byte(`{ "ram" : "` + pertRes + `" }`), scraper.out[4:])
            }

        case "net" :
            log.Print("Starting to get network data")
            res := scraper.getNetworkData(netIface)
            igPertRes := transBandwidthUnit(res[1])
            egPertRes := transBandwidthUnit(res[3])
            printResult(name, "Ingress", transBandwidthUnit(res[0]), igPertRes, percentile)
            printResult(name, "Egress", transBandwidthUnit(res[2]), egPertRes, percentile)

            if scraper.out[:4] == "api:" {
                log.Println("Calling API!")
                sendMetric([]byte(`{ "ingress" : "` + igPertRes + `", "egress" : "` + egPertRes + `" }`), scraper.out[4:])
            }

        case "all":
            log.Print("Starting to get all metrics: ")
            cpuRes, memRes, netRes := scraper.getAllData(netIface)

            cpuPertRes := transCpuUnitV2(cpuRes[1])
            printResult(name, "CPU", transCpuUnitV2(cpuRes[0]), cpuPertRes, percentile)

            memPertRes := transMemoryUnit(memRes[1])
            printResult(name, "RAM", transMemoryUnit(memRes[0]), memPertRes, percentile)

            igPertRes := transBandwidthUnit(netRes[1])
            egPertRes := transBandwidthUnit(netRes[3])
            printResult(name, "Ingress", transBandwidthUnit(netRes[0]), igPertRes, percentile)
            printResult(name, "Egress", transBandwidthUnit(netRes[2]), egPertRes, percentile)

            if scraper.out[:4] == "api:" {
                log.Println("Calling API!")
                sendMetric([]byte(`{ "cpu": "` + cpuPertRes +
                                 `", "ram": "` + memPertRes +
                                 `", "ingress": "` + igPertRes +
                                 `", "egress": "` + egPertRes + `" }`), scraper.out[4:])
            }


        default:
            log.Fatal("metric type is not in the handling list")
    }

    log.Print("Colibri is successfully completed !")

}
