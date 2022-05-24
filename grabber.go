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
    "time"
    "flag"
    "fmt"
    "log"
    "io/ioutil"
    "strings"
)

func findEth0Index(data string) int {
//    data_list = strings.Split(data, "\n")
    for i, d := range strings.Split(data, "\n") {
        if strings.Contains(d, "eth0") {
            return i
        }
    }
    return -1
}

func scrapeNetMetrics(interval int, metrics_path string) []string {

    var outputs [iterates]string
}

func main() {

    var pid string
    var interval_ms int
    const iterates = 10

    flag.StringVar(&pid, "pid", "0", "The process ID of the container.")
    flag.IntVar(&interval_ms, "freq", 5, "The scraping time of metrics collection in millisecond.")
    flag.Parse()

    if interval_ms <= 0 {
        log.Print("Monitoring process cannot be processed with interval_ms less and equal 0.")
        return
    }

    //%% print starting time

    var outputs [iterates]string
    for i := 0; i < iterates; i++ {
        net_stat, err := ioutil.ReadFile("/tmp/proc/"+pid+"/net/dev")

        if err != nil {
            log.Fatal(err)
        }
        outputs[i] = string(net_stat)
        time.Sleep(time.Duration(interval_ms) * time.Millisecond)
    }

    eth0_idx := findEth0Index(outputs[0])

    if eth0_idx < 0 {
        log.Print("No eth0's info.")
        return
    }

    for i := 0; i < iterates; i++ {
        metrics := strings.Fields(strings.Split(outputs[i],"\n")[eth0_idx])
        fmt.Printf("%s %s %s %s \n", metrics[1], metrics[2], metrics[9], metrics[10])
    }
}
