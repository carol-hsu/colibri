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
    "io/ioutil"
    "strings"
    "strconv"
    "github.com/montanaflynn/stats"
)

const percent = 95

func stringToInt(str string) int {
    n, _ := strconv.Atoi(str)
    return n
}

func getCgroupMetricPath(cgroup_path string, keyword string) string {

    file_content, err := ioutil.ReadFile(cgroup_path)

    if err != nil {
        fmt.Printf("Error: %v\n", err)
    } else {
        for _, path := range strings.Split(string(file_content), "\n") {
            if strings.Contains(path, keyword) {
                return strings.Split(path, ":")[2]
            }
        }
    }
    return ""

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

func countRate(data []string, interval int) []float64 {

    res := make([]float64, 2)
    float_data := make([]float64, len(data)-1)

    for i := 0; i < len(data)-1; i++ {
        float_data[i] = float64((stringToInt(data[i+1]) - stringToInt(data[i])) / interval)
    }

    res[0], _ = stats.Mean(float_data)
    res[1], _ = stats.Percentile(float_data, 95)

    return res
}

func countValue(data []float64) []float64 {

    res := make([]float64, 2)
    res[0], _ = stats.Mean(data)
    res[1], _ = stats.Percentile(data, 95)
    return res
}

func transBandwidthUnit(bw float64) string {
    // change X/ms to Y/s, Y's minimum unit is k
    // check if the value fit for k or M
    if (bw*1000/1024) < 1 {
    // less than 1 KB, return minimal one
        return "1k"
    }else if (bw*1000/1024/1024) < 1 {
    // less than KB use "k"
        return fmt.Sprint(math.Round(bw*1000/1024))+"k"
    }else{
    // larger than MB, use "M"
        return fmt.Sprint(math.Round(bw*1000/1024/1024))+"M"
    }
}
