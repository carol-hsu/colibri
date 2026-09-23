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

package lib

import (
	"log"
	"os"
	"strings"
    "bufio"
    "errors"
)

const (

	// cgroup_path = "/sys/fs/cgroups/" could be replaced by below
	// To avoid mixing host's data to container(scraper)'s data, we will mount host data to /tmp
	//PidCgroupPath        = "/tmp/proc/{pid}/cgroup" //comes out the full path of CPU and RAM
	//NetMetricsPath       = "/tmp/proc/{pid}/net/dev"
	//CgroupFilesystemDir  = "/tmp/cgroup"
	//CgroupFilesystemPath = CgroupFilesystemDir + "/"

    rootSchedFile = "/proc/1/sched"
	CpuDirectory = "cpu,cpuacct"
	MemDirectory = "memory"
)

type PathFinder struct {
	// path of the yellow page to get the full path of container
	pidPath string
	// path of network metric
	netPath string
	// path of cgroupfs, the prefix of directory leading to find container metrics
	cgroupPath string
}

func CreatePathFinder(pid string) (*PathFinder, error) {
	// detect if the process is running on a host or in a container

    file, err := os.Open(rootSchedFile)
    
    if err != nil {
        return nil, err
    }
    
	defer file.Close()

    scanner := bufio.NewScanner(file)

    if scanner.Scan() {
	    if strings.Contains(scanner.Text(), "colibri") {
		    // we are in a container
		    return &PathFinder{"/tmp/proc/" + pid + "/cgroup",
                               "/tmp/proc/" + pid + "/net/dev", 
                               "/tmp/cgroup"},
			       nil
	    }

    }else {
        return nil, errors.New(rootSchedFile + " is empty.")
    }

	if err != nil {
		return nil, err
	}


	return &PathFinder{"/proc/" + pid + "/cgroup",
					   "/proc/" + pid + "/net/dev",
					   "/sys/fs/cgroup"},
		   nil
}

func (pf *PathFinder) getCgroupMetricPath(keyword string) string {

	content, err := os.ReadFile(pf.pidPath)

	if err != nil {
		log.Print("Cannot read cgroup metric path: ", err)
	} else if len(keyword) == 0 {
		// v2: return the first line, since it is the only line
		// remove all /../ relative path
		path := strings.TrimSpace(strings.Split(string(content), ":")[2])
		for strings.HasPrefix(path, "/..") {
			path = path[3:]
		}
		return path

	} else {
		for _, path := range strings.Split(string(content), "\n") {
			if strings.Contains(path, keyword) {
				return strings.Split(path, ":")[2]
			}
		}
	}
	return ""

}
/* TODO: need to fix to support v1
func GetCpuPath(pid string) string {

	path := getCgroupMetricPath(strings.Replace(PidCgroupPath, "{pid}", pid, 1), CpuDirectory)

	if path == "" {
		log.Fatal("Error: (cgroup v1) failed to find the path of CPU data\n")
	}

	return CgroupFilesystemPath + CpuDirectory + path + "/cpuacct.usage"
}
*/

func (pf *PathFinder) GetCpuPathV2() string {

	//path := getCgroupMetricPath(strings.Replace(PidCgroupPath, "{pid}", pid, 1), "")
	path := pf.getCgroupMetricPath("")

	if path == "" {
		log.Fatal("Error: (cgroup v2) failed to find the path of CPU data\n")
	}

	return pf.cgroupPath + "/" + path + "/cpu.stat"
}


/* TODO: need to fix to support v1
func GetMemPath(pid string) (string, string) {

	path := getCgroupMetricPath(strings.Replace(PidCgroupPath, "{pid}", pid, 1), MemDirectory)

	if path == "" {
		log.Fatal("Error: failed to find the path of Memory data\n")
	}

	return CgroupFilesystemPath + MemDirectory + path + "/memory.usage_in_bytes",
		CgroupFilesystemPath + MemDirectory + path + "/memory.stat"

}
*/

func (pf *PathFinder) GetMemPathV2() (string, string) {

	path := pf.getCgroupMetricPath("")

	if path == "" {
		log.Fatal("Error: failed to find the path of Memory data\n")
	}

	return pf.cgroupPath + "/" + path + "/memory.current",
		pf.cgroupPath + "/" + path + "/memory.stat"

}

func (pf *PathFinder) GetNetPath() string {
	//cgroup v1 and v2 use the same path for network numbers
	return pf.netPath

}
