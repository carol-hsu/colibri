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
    "io/ioutil"
    "strings"
    "strconv"
)

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
