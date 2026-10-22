// Copyright 2026 Carol Hsu
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

package metrics

import (
    "time"
    "fmt"
    "log"
    "os"
    "strings"

	utils "colibri/internal/lib"
)

func GetCpuMetrics(path string, iteration int, interval int, ch chan<-[]string) []string {


    var statsOutput = []string{}
    for i:=0; i<iteration; i++ {
        t0 := time.Now()
        stats, err  := os.ReadFile(path)
        if err != nil {
            if i == 0 {
            // nothing existed in output, then forcefully stop
                log.Fatal(err)
            }else{
                log.Print("App stopped earlier, starting to print output")
                break
            }
        }
        t1 := time.Now()
        
        statsOutput = append(statsOutput, string(stats))
        time.Sleep(time.Duration(interval) * time.Millisecond)
        t2 := time.Now()

        ioDura := t1.Sub(t0)
        interDura := t2.Sub(t1)
        fmt.Println(ioDura.Nanoseconds(), interDura.Nanoseconds())
    }

    if ch != nil {
        ch <- statsOutput
    }

    return statsOutput
}


func AnalyzeCpuMetrics(stats []string, filePrefix string, interval int, percent float64, ch chan<-[]float64) []float64 {

    var outputs = make([]string, len(stats))

    usageIdx := utils.FindIndex(stats[0], "usage_usec")

    for i := 0; i < len(outputs); i++ {
        outputs[i] = strings.Fields(strings.Split(stats[i], "\n")[usageIdx])[1]
    }

    //if filePrefix is empty, then don't write out, just print analysis result
    if len(filePrefix) > 0 {
        file := utils.CreateOutputFile(filePrefix + "_" +fmt.Sprint(interval) + "ms_cpu")
        defer file.Close()

        for i:=0; i<len(outputs); i++ {
            file.WriteString(outputs[i]+"\n")
        }
    }

    rate := utils.CountRate(outputs, interval, percent)

    if ch != nil {
        ch <- rate
    }

    return rate
}


func getCpuValue(path string, idx int) string {

    stats, err  := os.ReadFile(path)
    if err != nil {
        log.Print("Cannot read statistic file of cpu: ", err)
        return ""
    }

    return strings.Fields(strings.Split(string(stats), "\n")[idx])[1]
}

