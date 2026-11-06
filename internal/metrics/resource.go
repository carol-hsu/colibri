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
    "strconv"
)

func GetMetrics(path string, iteration int, interval int, outputFile string, ch chan<-[]string) []string {

    statsOutput, ioTime, sleepTime := []string{}, []time.Duration{}, []time.Duration{}

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

        ioTime = append(ioTime, t1.Sub(t0))
        sleepTime = append(sleepTime, t2.Sub(t1))
    }

    if ch != nil {
        ch <- statsOutput
    }
    StoreScrapeDelay(outputFile, ioTime, sleepTime)

    return statsOutput
}


func StoreScrapeDelay(filename string, ioDuration []time.Duration, sleepDuration []time.Duration) {

    f, err := os.Create(filename)
    defer f.Close()

    if err == nil {
        for i:=0; i<len(ioDuration); i++ {
            ioTime := strconv.FormatInt(ioDuration[i].Nanoseconds(), 10)
            sleepTime := strconv.FormatInt(sleepDuration[i].Nanoseconds(), 10)
            f.WriteString(ioTime+" "+sleepTime+"\n")
        }
    } else {
        log.Println("Failed to create file [", filename ,"] to output scraping delay:", err)
    }

}
