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


func GetMemoryMetrics(usagePath string, statsPath string, iteration int, interval int, ch chan<-[]string) ([]string, []string) {

    var usageOutput, statsOutput = []string{}, []string{}

    for i:=0; i<iteration; i++ {
        t0 := time.Now()
        usage, err  := os.ReadFile(usagePath)
        if err != nil {
            if i == 0 {
            // nothing existed in output, then forcefully stop
                log.Fatal(err)
            }else{
                log.Print("Monitored application is stopped, starting to print output")
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
                log.Print("Monitored application is stopped, starting to print output")
                break
            }
        }
        statsOutput = append(statsOutput, string(stats))

        t1 := time.Now()
        time.Sleep(time.Duration(interval) * time.Millisecond)
        t2 := time.Now()

        ioDura := t1.Sub(t0)
        interDura := t2.Sub(t1)
        fmt.Println("Memory", ioDura.Nanoseconds(), interDura.Nanoseconds())
    }

    if ch != nil {
        ch <- usageOutput
        ch <- statsOutput
    }

    return usageOutput, statsOutput

}

func AnalyzeMemoryMetrics(usage []string, stats []string, filename string, percent float64, ch chan<-[]float64) []float64 {
    
    metricSize := len(stats)
    var outputs = make([]float64, metricSize)

    inactiveFileIdx := utils.FindIndex(stats[0], "inactive_file")

    for i:=0; i<metricSize; i++ {
        outputs[i] = utils.StringToFloat(usage[i]) - utils.StringToFloat(strings.Fields(strings.Split(stats[i], "\n")[inactiveFileIdx])[1])
    }

    //if filename is none, then don't write out, just print analysis result
    if len(filename) > 0{
        file := utils.CreateOutputFile(filename)
        defer file.Close()

        for i := 0; i < len(outputs); i++ {
            file.WriteString(fmt.Sprintf("%f\n", outputs[i]))
        }
    }

    value := utils.CountValue(outputs, percent)

    if ch != nil {
        ch <- value
    }

    return value

}
