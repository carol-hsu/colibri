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
// See the License for the specific language s.verning permissions and
// limitations under the License.

package main

import (
    "bytes"
    "io/ioutil"
    "log"
    "net/http"
    "crypto/x509"
    "crypto/tls"
)

const (
    Url = "https://10.96.0.1/api/v1/namespaces/colibri/services/colibri-apiserver:http/proxy/colibri/"
    TokenFile = "/var/run/secrets/kubernetes.io/serviceaccount/token"
    CaFile = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
)

func createHttpClient() *http.Client {

    caCert, err := ioutil.ReadFile(CaFile)
    if err != nil {
        log.Fatal(err)
    }
    caCertPool := x509.NewCertPool()
    caCertPool.AppendCertsFromPEM(caCert)

    client := &http.Client{
        Transport: &http.Transport{
            TLSClientConfig: &tls.Config{
                RootCAs:      caCertPool,
            },
        },
    }
    return client
}

func sendMetric(value []byte, rid string) {

    //create bearer with token
    token, err := ioutil.ReadFile(TokenFile)
    if err != nil {
        log.Fatal(err)
    }

    var bearer = "Bearer " + string(token)

    //build request
    req, err := http.NewRequest("POST", Url+rid, bytes.NewBuffer(value))
    if err != nil {
        log.Fatal(err)
    }

    // add headers
    req.Header.Add("Authorization", bearer)
    req.Header.Add("Content-Type", "application/json")

    // create http client
    client := createHttpClient()

    resp, err := client.Do(req)

    if err != nil {
        log.Fatal(err)
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Println("Error while reading the response bytes:", err)
    }
    log.Println(string([]byte(body)))
}
