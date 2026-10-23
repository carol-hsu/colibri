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
    "strings"

	utils "colibri/internal/lib"
)

func AnalyzeNetworkMetrics(stats []string, idx int, filePrefix string, interval int, percent float64, ch chan<-[]float64) []float64 {
    
    //parse bandwidth value, and store separately
    var outputLen = len(stats)
    var igBw = make([]string, outputLen)
    var egBw = make([]string, outputLen)

    for i := 0; i < outputLen; i++ {
        metrics := strings.Fields(strings.Split(stats[i], "\n")[idx])
        igBw[i] = metrics[1]
        egBw[i] = metrics[9]
    }

    //if filePrefix is nil, then don't write out, just print analysis result
    if len(filePrefix) > 0 {
        igFile := utils.CreateOutputFile(filePrefix + "ig_bytes")
        egFile := utils.CreateOutputFile(filePrefix + "eg_bytes")

        defer igFile.Close()
        defer egFile.Close()

        // create output files
        for i := 0; i < outputLen; i++ {
            igFile.WriteString(igBw[i]+"\n")
            egFile.WriteString(egBw[i]+"\n")
        }
    }

    rate := append(utils.CountRate(igBw, interval, percent), utils.CountRate(egBw, interval, percent)...)

    if ch != nil {
        ch <- rate
    }

    return rate

}
