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
    "flag"
    "fmt"
    "log"
    "os"
    "strings"

    utils "colibri/internal/lib"
    metrics "colibri/internal/metrics"
)

type Scraper struct {
	// The manager of file path configuration
	pf *utils.PathFinder
    // The postfix name of output file
    out string
    // The metric scraping timespan in millisecond
    ms int
    // The scrapping #iteration
    iter int
    // The percentile for output
    pert float64
}

const logPath = "/log/"

func (s Scraper) getCpuData() []float64 {

    path := s.pf.GetCpuPathV2()
    stats := metrics.GetMetrics("CPU", path, s.iter, s.ms, nil)

    log.Print("CPU metrics collection is finished. Start to post-process data ...")
    var output = []float64{}

    if strings.Contains(s.out, "file:") {
        output = metrics.AnalyzeCpuMetrics(stats, logPath+s.out[5:]+"_"+fmt.Sprint(s.ms)+"ms_cpu", s.ms, s.pert, nil)
    } else {
        output = metrics.AnalyzeCpuMetrics(stats, "", s.ms, s.pert, nil)
    }

    return output
}

func (s Scraper) getMemoryData() []float64 {

    usagePath, statsPath := s.pf.GetMemPathV2()
    usage, stats := metrics.GetMemoryMetrics(usagePath, statsPath, s.iter, s.ms, nil)

    log.Print("Memory metrics collection is finished. Start to post-process data ...")
    var output = []float64{}

    if strings.Contains(s.out, "file:") {
        output = metrics.AnalyzeMemoryMetrics(usage, stats, logPath+s.out[5:]+"_"+fmt.Sprint(s.ms)+"ms_mem", s.pert, nil)
    } else {
        output = metrics.AnalyzeMemoryMetrics(usage, stats, "", s.pert, nil)
    }

    return output
}

func (s Scraper) getNetworkData(iface string) []float64 {

    path := s.pf.GetNetPath()
    //get index for collecting data from memory statistic file
    ifaceIdx := getIfaceIndex(path, iface)
    if ifaceIdx < 0 {
        log.Fatal("The requested network interface is not existed:", iface)
    }

    stats := metrics.GetMetrics("Network", path, s.iter, s.ms, nil)

    log.Print("Network metrics collection is finished. Start to post-process data ...")

    var output = []float64{}

    if strings.Contains(s.out, "file:") {
        output = metrics.AnalyzeNetworkMetrics(stats, ifaceIdx, logPath+s.out[5:]+"_"+fmt.Sprint(s.ms)+"ms_", s.ms, s.pert, nil)
    } else {
        output = metrics.AnalyzeNetworkMetrics(stats, ifaceIdx, "", s.ms, s.pert, nil)
    }

    return output
}

func getIfaceIndex(file string, iface string) int {

    stats, err  := os.ReadFile(file)
    if err != nil {
        log.Print("Cannot read statistic file of network.")
        return -1
    }

    return utils.FindIndex(string(stats), iface)
}

func (s Scraper) getAllData(iface string) ([]float64, []float64, []float64) {

    cpuMetricChan := make(chan []string)
    cpuResultChan := make(chan []float64)
    memMetricChan := make(chan []string)
    memResultChan := make(chan []float64)
    netMetricChan := make(chan []string)
    netResultChan := make(chan []float64)

    //get statistic path of container in cgroupfs
    cpuPath := s.pf.GetCpuPathV2()
    netPath := s.pf.GetNetPath()
    usagePath, statsPath := s.pf.GetMemPathV2()

    //get index for collecting data from memory statistic file
    ifaceIdx := getIfaceIndex(netPath, iface)
    if ifaceIdx < 0 {
        log.Fatal("The requested network interface is not existed:", iface)
    }

    go metrics.GetMetrics("CPU", cpuPath, s.iter, s.ms, cpuMetricChan)
    go metrics.GetMetrics("Network", netPath, s.iter, s.ms, netMetricChan)
    go metrics.GetMemoryMetrics(usagePath, statsPath, s.iter, s.ms, memMetricChan)

    cpuStats := <- cpuMetricChan
    memUsage := <- memMetricChan
    memStats := <- memMetricChan
    netStats := <- netMetricChan

    //if outputName == none, then don't write out, just print analysis result
    if strings.Contains(s.out, "file:") {

        filePrefix := logPath + s.out[5:] + "_" +fmt.Sprint(s.ms)+"ms_"
        
        go metrics.AnalyzeCpuMetrics(cpuStats, filePrefix+"cpu", s.ms, s.pert, cpuResultChan)
        go metrics.AnalyzeMemoryMetrics(memUsage, memStats, filePrefix+"mem", s.pert, memResultChan)
        go metrics.AnalyzeNetworkMetrics(netStats, ifaceIdx, filePrefix, s.ms, s.pert, netResultChan)

    }else {
        go metrics.AnalyzeCpuMetrics(cpuStats, "", s.ms, s.pert, cpuResultChan)
        go metrics.AnalyzeMemoryMetrics(memUsage, memStats, "", s.pert, memResultChan)
        go metrics.AnalyzeNetworkMetrics(netStats, ifaceIdx, "", s.ms, s.pert, netResultChan)
    }

    cpuRes := <-cpuResultChan
    memRes := <-memResultChan
    netRes := <-netResultChan

    return cpuRes, memRes, netRes
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
	//initialize the path config
	pf, err := utils.CreatePathFinder(pid)

	if err != nil {
        log.Printf("Failed to create a PathFinder for target container: ", err)
        return
	}

    scraper := Scraper{pf, outputName, intervalMsec, iterateNum, percentile}

    //getting numbers by type
    switch metricType {
        case "cpu" :
            log.Print("Starting to get CPU data")
            res := scraper.getCpuData()
            pertRes := utils.TransCpuUnitV2(res[1])
            utils.PrintResult(name, "CPU", utils.TransCpuUnitV2(res[0]), pertRes, percentile)

            if scraper.out[:4] == "api:" {
                log.Println("Calling API!")
                utils.SendMetric([]byte(`{ "cpu" : "` + pertRes + `" }`), scraper.out[4:])
            }

        case "mem" :
            log.Print("Starting to get RAM data")
            res := scraper.getMemoryData()
            pertRes := utils.TransMemoryUnit(res[1])
            utils.PrintResult(name, "RAM", utils.TransMemoryUnit(res[0]), pertRes, percentile)

            if scraper.out[:4] == "api:" {
                log.Println("Calling API!")
                utils.SendMetric([]byte(`{ "ram" : "` + pertRes + `" }`), scraper.out[4:])
            }

        case "net" :
            log.Print("Starting to get network data")
            res := scraper.getNetworkData(netIface)
            igPertRes := utils.TransBandwidthUnit(res[1])
            egPertRes := utils.TransBandwidthUnit(res[3])
            utils.PrintResult(name, "Ingress", utils.TransBandwidthUnit(res[0]), igPertRes, percentile)
            utils.PrintResult(name, "Egress", utils.TransBandwidthUnit(res[2]), egPertRes, percentile)

            if scraper.out[:4] == "api:" {
                log.Println("Calling API!")
                utils.SendMetric([]byte(`{ "ingress" : "` + igPertRes + `", "egress" : "` + egPertRes + `" }`), scraper.out[4:])
            }

        case "all":
            log.Print("Starting to get all metrics: ")
            cpuRes, memRes, netRes := scraper.getAllData(netIface)

            cpuPertRes := utils.TransCpuUnitV2(cpuRes[1])
            utils.PrintResult(name, "CPU", utils.TransCpuUnitV2(cpuRes[0]), cpuPertRes, percentile)

            memPertRes := utils.TransMemoryUnit(memRes[1])
            utils.PrintResult(name, "RAM", utils.TransMemoryUnit(memRes[0]), memPertRes, percentile)

            igPertRes := utils.TransBandwidthUnit(netRes[1])
            egPertRes := utils.TransBandwidthUnit(netRes[3])
            utils.PrintResult(name, "Ingress", utils.TransBandwidthUnit(netRes[0]), igPertRes, percentile)
            utils.PrintResult(name, "Egress", utils.TransBandwidthUnit(netRes[2]), egPertRes, percentile)

            if scraper.out[:4] == "api:" {
                log.Println("Calling API!")
                utils.SendMetric([]byte(`{ "cpu": "` + cpuPertRes +
                                 `", "ram": "` + memPertRes +
                                 `", "ingress": "` + igPertRes +
                                 `", "egress": "` + egPertRes + `" }`), scraper.out[4:])
            }


        default:
            log.Fatal("metric type is not in the handling list")
    }

    log.Print("Colibri is successfully completed !")

}
