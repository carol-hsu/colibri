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
    "os"
    "strings"
    "strconv"

	utils "colibri/internal/lib"
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

    cpuPath := s.pf.GetCpuPathV2()
    var statsOutput = []string{}

    for i:=0; i<s.iter; i++ {
        t0 := time.Now()
        stats, err  := os.ReadFile(cpuPath)
        if err != nil {
            if i == 0 {
            // nothing existed in output, then forcefully stop
                log.Fatal(err)
            }else{
                log.Print("App stopped earlier, starting to print output")
                break
            }
        }
        statsOutput = append(statsOutput, string(stats))
        time.Sleep(time.Duration(s.ms) * time.Millisecond)
        dura := time.Now().Sub(t0)
        fmt.Println(dura.Nanoseconds())
    }

    log.Print("CPU metrics collection is finished. Start to post-process data ...")

    var outputs = make([]string, len(statsOutput))

    usageIdx := utils.FindIndex(statsOutput[0], "usage_usec")

    for i := 0; i < len(outputs); i++ {
        outputs[i] = strings.Fields(strings.Split(statsOutput[i], "\n")[usageIdx])[1]
    }

    //if outputName == none, then don't write out, just print analysis result
    if strings.Contains(s.out, "file:") {
        file := utils.CreateOutputFile(logPath + s.out[5:] + "_" +fmt.Sprint(s.ms) + "ms_cpu")
        defer file.Close()

        for i:=0; i<len(outputs); i++ {
            file.WriteString(outputs[i]+"\n")
        }
    }

    return utils.CountRate(outputs, s.ms, s.pert)
}

func (s Scraper) getMemoryData() []float64 {

    usagePath, statsPath := s.pf.GetMemPathV2()

    var usageOutput, statsOutput = []string{}, []string{}

    for i:=0; i < s.iter; i++ {

        usage, err  := os.ReadFile(usagePath)
        if err != nil {
            if i == 0 {
            // nothing existed in output, then forcefully stop
                log.Fatal(err)
            }else{
                log.Print("App stopped earlier, starting to print output")
                break
            }
        }
        usageOutput = append(usageOutput, strings.TrimSpace(string(usage)))

        stats, err  := os.ReadFile(statsPath)
        if err != nil {
            if i == 0 {
            // nothing existed in output, then forcefully stop
                log.Fatal(err)
            }else{
                log.Print("App stopped earlier, starting to print output")
                break
            }
        }
        statsOutput = append(statsOutput, string(stats))

        time.Sleep(time.Duration(s.ms) * time.Millisecond)
    }

    log.Print("Memory metrics collection is finished. Start to post-process data ...")

    //count usage and inactive file size, and stored in float
    var outputs = make([]float64, len(statsOutput))

    inactiveFileIdx := utils.FindIndex(statsOutput[0], "inactive_file")

    for i := 0; i < len(outputs); i++ {
        v := utils.StringToFloat(usageOutput[i]) - utils.StringToFloat(strings.Fields(strings.Split(statsOutput[i], "\n")[inactiveFileIdx])[1])
        outputs[i] = v
    }

    //if outputName == none, then don't write out, just print analysis result
    if strings.Contains(s.out, "file:") {
        file := utils.CreateOutputFile(logPath + s.out[5:] + "_" +fmt.Sprint(s.ms) + "ms_mem")
        defer file.Close()

        for i := 0; i < len(outputs); i++ {
            file.WriteString(fmt.Sprintf("%f\n", outputs[i]))
        }
    }

    return utils.CountValue(outputs, s.pert)
}

func (s Scraper) getNetworkData(iface string) []float64 {

    var outputs =[]string{}
    path := s.pf.GetNetPath()

    for i := 0; i < s.iter; i++ {
        net_stat, err := os.ReadFile(path)

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

    ifaceIdx := utils.FindIndex(outputs[0], iface)

    if ifaceIdx < 0 {
        log.Fatal("No info for the specified interface")
    }

    //parse bandwidth value, and store separately
    var outputLen = len(outputs)
    var igBw = make([]string, outputLen)
    var egBw = make([]string, outputLen)

    for i := 0; i < outputLen; i++ {
        metrics := strings.Fields(strings.Split(outputs[i], "\n")[ifaceIdx])
        igBw[i] = metrics[1]
        egBw[i] = metrics[9]
    }

    //if outputName == none, then don't write out, just print analysis result
    if strings.Contains(s.out, "file:") {
        igFile := utils.CreateOutputFile(logPath + s.out[5:] + "_" + fmt.Sprint(s.ms) + "ms_ig_bytes")
        egFile := utils.CreateOutputFile(logPath + s.out[5:] + "_" + fmt.Sprint(s.ms) + "ms_eg_bytes")

        defer igFile.Close()
        defer egFile.Close()

        // create output files
        for i := 0; i < outputLen; i++ {
            igFile.WriteString(igBw[i]+"\n")
            egFile.WriteString(egBw[i]+"\n")
        }
    }

    igRes := utils.CountRate(igBw, s.ms, s.pert)
    egRes := utils.CountRate(egBw, s.ms, s.pert)


    return append(igRes, egRes...)
}

func getCpuValue(path string, idx int) string {

    stats, err  := os.ReadFile(path)
    if err != nil {
        log.Print("Cannot read statistic file of cpu: ", err)
        return ""
    }

    return strings.Fields(strings.Split(string(stats), "\n")[idx])[1]
}

func getMemoryValue(usageFile string, statsFile string, idx int) float64 {

    usage, err  := os.ReadFile(usageFile)
    if err != nil {
        log.Print("Cannot read usage file of memory: ", err)
        return -1
    }

    usageOutput := strings.TrimSpace(string(usage))

    stats, err  := os.ReadFile(statsFile)
    if err != nil {
        log.Print("Cannot read statistic file of memory: ", err)
        return -1
    }

    return utils.StringToFloat(usageOutput) - utils.StringToFloat(strings.Fields(strings.Split(string(stats), "\n")[idx])[1])
}

func getUsageIndex(file string) int {

    stats, err  := os.ReadFile(file)
    if err != nil {
        log.Print("Cannot read statistic file of cpu: ", err)
        return -1
    }

    return utils.FindIndex(string(stats), "usage_usec")
}


func getInactiveFileIndex(file string) int {

    stats, err  := os.ReadFile(file)
    if err != nil {
        log.Print("Cannot read statistic file of memory: ", err)
        return -1
    }

    return utils.FindIndex(string(stats), "inactive_file")
}

func getNetworkValue(file string, idx int) (string, string) {

    netStats, err := os.ReadFile(file)
    if err != nil {
        log.Print("Cannot read statistic file of network.")
        return "", ""
    }
    stats := strings.Fields(strings.Split(string(netStats), "\n")[idx])

    return stats[1], stats[9]
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
    //get path of container
    cpuPath := s.pf.GetCpuPathV2()
    usagePath, statsPath := s.pf.GetMemPathV2()
    netPath := s.pf.GetNetPath()

    var cpuOutput, igOutput, egOutput, timeOutput []string
    var memOutput = []float64{}

    //get index for collecting data from memory statistic file
    cpuIdx := getUsageIndex(cpuPath)
    memIdx := getInactiveFileIndex(statsPath)
    netIdx := getIfaceIndex(netPath, iface)

    //start metrics scraping period
    for i:=0; i < s.iter; i++ {
        t0 := time.Now()
        cpuVal := getCpuValue(cpuPath, cpuIdx)
        memVal := getMemoryValue(usagePath, statsPath, memIdx)
        igBw, egBw := getNetworkValue(netPath, netIdx)

        if memVal < 0 || len(cpuVal) == 0 || len(igBw) == 0 {
            log.Print("App stopped earlier, starting to print output")
            break
        }

        cpuOutput = append(cpuOutput, cpuVal)
        memOutput = append(memOutput, memVal)
        igOutput = append(igOutput, igBw)
        egOutput = append(egOutput, egBw)

        time.Sleep(time.Duration(s.ms) * time.Millisecond)
        dura := time.Now().Sub(t0)
        timeOutput = append(timeOutput, strconv.Itoa(int(dura.Nanoseconds())))
    }

    //if outputName == none, then don't write out, just print analysis result
    if strings.Contains(s.out, "file:") {
        filePrefix := logPath + s.out[5:] + "_" +fmt.Sprint(s.ms)

        cpuFile := utils.CreateOutputFile(filePrefix + "ms_cpu")
        defer cpuFile.Close()

        memFile := utils.CreateOutputFile(filePrefix + "ms_mem")
        defer memFile.Close()

        igFile := utils.CreateOutputFile(filePrefix + "ms_ig_bytes")
        defer igFile.Close()

        egFile := utils.CreateOutputFile(filePrefix + "ms_eg_bytes")
        defer egFile.Close()

        timeFile := utils.CreateOutputFile(filePrefix + "ms_intervals")
        defer timeFile.Close()
        // TODO: can write become more efficient?
        for i := 0; i < len(cpuOutput); i++ {
            cpuFile.WriteString(cpuOutput[i]+"\n")
            memFile.WriteString(fmt.Sprintf("%.0f\n", memOutput[i]))
            igFile.WriteString(igOutput[i]+"\n")
            egFile.WriteString(egOutput[i]+"\n")
            timeFile.WriteString(timeOutput[i]+"\n")
        }
    }

    cpuRes := utils.CountRate(cpuOutput, s.ms, s.pert)
    memRes := utils.CountValue(memOutput, s.pert)
    igRes := utils.CountRate(igOutput, s.ms, s.pert)
    egRes := utils.CountRate(egOutput, s.ms, s.pert)

    return cpuRes, memRes, append(igRes, egRes...)
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
        //log.Print("Test done!")
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
