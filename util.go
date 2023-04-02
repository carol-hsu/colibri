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
    "log"
    "os"
    "math"
    "strings"
    "strconv"
    "github.com/montanaflynn/stats"
)

func stringToFloat(str string) float64 {
    f, _ := strconv.ParseFloat(str, 64)
    return f
}

// to find which row of data contains the keyword, which would be the information we want
func findIndex(data string, keyword string) int {
    for i, d := range strings.Split(data, "\n") {
        if strings.Contains(d, keyword) {
            return i
        }
    }
    return -1
}

//no help to close the file
func createOutputFile(filename string) *os.File {

    f, err := os.Create(filename)
    if err != nil {
        log.Fatal(err)
    }

    return f
}

func countRate(data []string, interval int, percent float64) []float64 {

    res := make([]float64, 2)
    float_data := make([]float64, len(data)-1)

    for i := 0; i < len(data)-1; i++ {
        float_data[i] = (stringToFloat(data[i+1]) - stringToFloat(data[i])) / float64(interval)
    }

    res[0], _ = stats.Mean(float_data)
    res[1], _ = stats.Percentile(float_data, percent)

    return res
}

func countValue(data []float64, percent float64) []float64 {

    res := make([]float64, 2)
    res[0], _ = stats.Mean(data)
    res[1], _ = stats.Percentile(data, percent)
    return res
}

func transCpuUnit(cpu float64) string {
    return strconv.Itoa(int(math.Round(cpu/1000)))+"m"
}

func transCpuUnitV2(cpu float64) string {
    return strconv.Itoa(int(math.Round(cpu)))+"m"
}

func transMemoryUnit(ram float64) string {
    return strconv.Itoa(int(math.Round(ram/1024/1024)))+"Mi"
}

func transBandwidthUnit(bw float64) string {
    // change X/ms to Y/s, Y's minimum unit is k
    // check if the value fit for k or M
    if (bw*1000/1024) < 1 {
    // less than 1 KB, return minimal one
          return fmt.Sprint(math.Round(bw*1000))
    //    return "1k"
    }else if (bw*1000/1024/1024) < 1 {
    // less than KB use "k"
        return fmt.Sprint(math.Round(bw*1000/1024))+"k"
    }else{
    // larger than MB, use "M"
        return fmt.Sprint(math.Round(bw*1000/1024/1024))+"M"
    }
}

func printResult(workName string, metricName string, avgValue string, pertValue string, pert float64) {

    log.Printf("%s -- %s Avg: %s, %.2f-Percentile: %s\n", workName, metricName, avgValue, pert, pertValue)

}
