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
    "log"
    "os"
    "fmt"
    "strings"
    "sync"

	utils "colibri/internal/lib"
)

func GetMemoryMetrics(usagePath string, statsPath string, iteration int, interval int, ch chan<-[]string) ([]string, []string) {

    usageOutput, statsOutput := []string{}, []string{}

    var usage, stats []byte
    var usageErr, statsErr error

    var wg sync.WaitGroup

    for i:=0; i<iteration; i++ {

        wg.Add(2)

        go func(){
            defer wg.Done()
            usage, usageErr = os.ReadFile(usagePath)
        }()

        go func(){
            defer wg.Done()
            stats, statsErr = os.ReadFile(statsPath)
        }()

        wg.Wait()

        if usageErr != nil || statsErr != nil {
            if i == 0 {
                //noting in the output; the file is not existed in the beginning;
                // TODO: do a better log handling here
                if usageErr == nil {
                    log.Fatal(usageErr)
                } else {
                    log.Fatal(statsErr)
                }
            } else {
                log.Print("Monitored application is stopped, starting to print output")
                break
            }
        }
        usageOutput = append(usageOutput, strings.TrimSpace(string(usage)))
        statsOutput = append(statsOutput, string(stats))

        time.Sleep(time.Duration(interval) * time.Millisecond)
    }

    if ch != nil {
        ch <- usageOutput
        ch <- statsOutput
    }

    return usageOutput, statsOutput

}

func GetMemoryMetricsDelayTest(usagePath string, statsPath string, iteration int, interval int, outputFile string, ch chan<-[]string) ([]string, []string) {

    usageOutput, statsOutput := []string{}, []string{}
    ioTime, sleepTime := []time.Duration{}, []time.Duration{}

    var usage, stats []byte
    var usageErr, statsErr error

    var wg sync.WaitGroup

    for i:=0; i<iteration; i++ {

        t0 := time.Now()

        wg.Add(2)

        go func(){
            defer wg.Done()
            usage, usageErr = os.ReadFile(usagePath)
        }()

        go func(){
            defer wg.Done()
            stats, statsErr = os.ReadFile(statsPath)
        }()

        wg.Wait()

        if usageErr != nil || statsErr != nil {
            if i == 0 {
                //noting in the output; the file is not existed in the beginning;
                // TODO: do a better log handling here
                if usageErr == nil {
                    log.Fatal(usageErr)
                } else {
                    log.Fatal(statsErr)
                }
            } else {
                log.Print("Monitored application is stopped, starting to print output")
                break
            }
        }

        usageOutput = append(usageOutput, strings.TrimSpace(string(usage)))
        statsOutput = append(statsOutput, string(stats))

        t1 := time.Now()
        time.Sleep(time.Duration(interval) * time.Millisecond)
        t2 := time.Now()

        ioTime = append(ioTime, t1.Sub(t0))
        sleepTime = append(sleepTime, t2.Sub(t1))
    }

    if ch != nil {
        ch <- usageOutput
        ch <- statsOutput
    }

    StoreScrapeDelay(outputFile, ioTime, sleepTime)

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
