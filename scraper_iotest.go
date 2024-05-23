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
    "strings"
//    "bufio"
    "bytes"
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

    f, err := openCpuFileV2(s.pid)
    if err != nil {
        fmt.Println("having errors!!")
        log.Fatal(err)
    }
    defer f.Close()
    var buf bytes.Buffer
    var stats_outputs = []string{}

    for i:=0; i<s.iter; i++ {
        t0 := time.Now()
        buf.ReadFrom(f)
        fmt.Println(buf.String())
//        sc := bufio.NewScanner(f)
//        sc.Scan()
//        fmt.Println(sc.Text())

//        stats, err  := os.ReadFile(cpu_data_fullpath)
//        if err != nil {
//            if i == 0 {
//            // nothing existed in output, then forcefully stop
//                log.Fatal(err)
//            }else{
//                log.Print("App stopped earlier, starting to print output")
//                break
//            }
//        }
        stats_outputs = append(stats_outputs, buf.String())
        time.Sleep(time.Duration(s.ms) * time.Millisecond)
        dura := time.Now().Sub(t0)
        fmt.Println(dura.Nanoseconds())
        buf.Reset()
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
//    return []float64{0.2, 0.2}
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

        default:
            log.Fatal("metric type is not in the handling list")
    }

    log.Print("Colibri is successfully completed !")

}
