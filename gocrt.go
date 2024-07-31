
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

const BASE_URL = "https://crt.sh/?q=%s&output=json"

var subdomains = make(map[string]struct{})
var wildcardSubdomains = make(map[string]struct{})

func main() {
	domain := flag.String("domain", "", "Specify Target Domain to get subdomains from crt.sh")
	recursive := flag.Bool("recursive", false, "Do recursive search for subdomains")
	wildcard := flag.Bool("wildcard", false, "Include wildcard in output")
	flag.Parse()

	if *domain == "" {
		fmt.Println("Usage: go run main.go [Options] use -h for help")
		os.Exit(1)
	}

	crtsh(*domain)

	for subdomain := range subdomains {
		fmt.Println(subdomain)
	}

	if *recursive {
		for wildcardSubdomain := range wildcardSubdomains {
			wildcardSubdomain = strings.Replace(wildcardSubdomain, "*.", "%25.", 1)
			crtsh(wildcardSubdomain)
		}
	}

	if *wildcard {
		for wildcardSubdomain := range wildcardSubdomains {
			fmt.Println(wildcardSubdomain)
		}
	}
}

func crtsh(domain string) {
	url := fmt.Sprintf(BASE_URL, domain)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error fetching data:", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	var jsondata []map[string]interface{}
	if err := json.Unmarshal(body, &jsondata); err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return
	}

	for _, record := range jsondata {
		nameValue, ok := record["name_value"].(string)
		if !ok {
			continue
		}
		names := strings.Split(nameValue, "\n")
		for _, name := range names {
			if strings.Contains(name, "*") {
				if _, exists := wildcardSubdomains[name]; !exists {
					wildcardSubdomains[name] = struct{}{}
				}
			} else {
				if _, exists := subdomains[name]; !exists {
					subdomains[name] = struct{}{}
				}
			}
		}
	}
}
