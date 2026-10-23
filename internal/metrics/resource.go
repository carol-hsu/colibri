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
)

func GetMetrics(resource string, path string, iteration int, interval int, ch chan<-[]string) []string {

    var statsOutput = []string{}
    for i:=0; i<iteration; i++ {
        t0 := time.Now()
        stats, err  := os.ReadFile(path)
        if err != nil {
            if i == 0 {
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
        fmt.Println(resource, ioDura.Nanoseconds(), interDura.Nanoseconds())
    }

    if ch != nil {
        ch <- statsOutput
    }

    return statsOutput
}

