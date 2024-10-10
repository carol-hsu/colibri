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
	"strconv"
	"strings"
	"sync"

	"golang.org/x/sys/unix"
)

const (

	// cgroup_path = "/sys/fs/cgroups/" could be replaced by below
	// To avoid mixing host's data to container(scraper)'s data, we will mount host data to /tmp
	PidCgroupPath        = "/tmp/proc/{pid}/cgroup" //comes out the full path of CPU and RAM
	NetMetricsPath       = "/tmp/proc/{pid}/net/dev"
	CgroupFilesystemDir  = "/tmp/cgroup"
	CgroupFilesystemPath = CgroupFilesystemDir + "/"

	CpuDirectory = "cpu,cpuacct"
	MemDirectory = "memory"
)

var (
	cgroupFd     int = -1
	prepOnce     sync.Once
	prepErr      error
	resolveFlags uint64
)

func openCpuFileV2(pid string) (*os.File, error) {
	// referring to the implementation of opencontainers/runc/libcontainer/cgroups/file.go

	path := getCpuPathV2(pid)
	mode := os.FileMode(0)

	trimPath := strings.TrimPrefix(path, CgroupFilesystemPath)
	if prepareOpenat2() != nil {
		log.Print("Warn: prepare for Openat2 error: ", prepErr)
		return nil, prepErr
	}
	fd, err := unix.Openat2(cgroupFd, trimPath,
		&unix.OpenHow{
			Resolve: resolveFlags,
			Flags:   uint64(unix.O_RDONLY) | unix.O_CLOEXEC,
			Mode:    uint64(mode),
		})
	if err != nil {
		err = &os.PathError{Op: "openat2", Path: path, Err: err}
		fdStr := strconv.Itoa(cgroupFd)
		fdDest, _ := os.Readlink("/tmp/proc/self/fd/" + fdStr)
		if fdDest != CgroupFilesystemDir {
			err = fmt.Errorf("cgroupFd %s unexpectedly opened to %s != %s: %w",
				fdStr, fdDest, CgroupFilesystemDir, err)
		}
		return nil, err
	}
	return os.NewFile(uintptr(fd), path), nil
}

func getCgroupMetricPath(cgroupPath string, keyword string) string {

	content, err := os.ReadFile(cgroupPath)

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

func getCpuPath(pid string) string {

	path := getCgroupMetricPath(strings.Replace(PidCgroupPath, "{pid}", pid, 1), CpuDirectory)

	if path == "" {
		log.Fatal("Error: (cgroup v1) failed to find the path of CPU data\n")
	}

	return CgroupFilesystemPath + CpuDirectory + path + "/cpuacct.usage"
}

func getCpuPathV2(pid string) string {

	path := getCgroupMetricPath(strings.Replace(PidCgroupPath, "{pid}", pid, 1), "")

	if path == "" {
		log.Fatal("Error: (cgroup v2) failed to find the path of CPU data\n")
	}

	return CgroupFilesystemDir + path + "/cpu.stat"
}

func getMemPath(pid string) (string, string) {

	path := getCgroupMetricPath(strings.Replace(PidCgroupPath, "{pid}", pid, 1), MemDirectory)

	if path == "" {
		log.Fatal("Error: failed to find the path of Memory data\n")
	}

	return CgroupFilesystemPath + MemDirectory + path + "/memory.usage_in_bytes",
		CgroupFilesystemPath + MemDirectory + path + "/memory.stat"

}

func getMemPathV2(pid string) (string, string) {

	path := getCgroupMetricPath(strings.Replace(PidCgroupPath, "{pid}", pid, 1), "")

	if path == "" {
		log.Fatal("Error: failed to find the path of Memory data\n")
	}

	return CgroupFilesystemPath + path + "/memory.current",
		CgroupFilesystemPath + path + "/memory.stat"

}

func getNetPath(pid string) string {
	//cgroup v1 and v2 use the same path for network numbers
	return strings.Replace(NetMetricsPath, "{pid}", pid, 1)

}
