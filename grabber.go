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
)

func main () {

    var pid int
    var interval_ms int

    flag.IntVar(&pid, "pid", 0, "The process ID of the container.")
    flag.IntVar(&interval_ms, "freq", 5, "The scraping time of metrics collection.")
    flag.Parse()

    for i:=0; i < 10; i++ {
        fmt.Printf("get data %d from %d\n", i, pid)
        time.Sleep(time.Duration(interval_ms) * time.Millisecond)
    }

}
