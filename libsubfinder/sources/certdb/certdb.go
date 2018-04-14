// 
// certdb.go : A CertDB Subdomain parser in golang
// Written By : @ice3man (Nizamul Rana)
// 
// Distributed Under MIT License
// Copyrights (C) 2018 Ice3man
//

package certdb

import (
	"io/ioutil"
	"fmt"
	"regexp"

	"subfinder/libsubfinder/helper"
)

// all subdomains found
var subdomains []string 

// Parser subdomains from SSL Certificate Information Page
func findSubdomains(link string, state *helper.State, channel chan []string) {
	var subdomainsfound []string

	resp, err := helper.GetHTTPResponse("https://certdb.com"+link, state.Timeout)
	if err != nil {
		channel <- subdomainsfound
		return
	}

    // Get the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		channel <- subdomainsfound
		return
	}

	src := string(body)

	SubdomainRegex, err := regexp.Compile("<a href=\"https://certdb.com/domain/(.*)\"  target='_blank' class='link-underlined' >")
	if err != nil {
		channel <- subdomainsfound
		return
	}

	match := SubdomainRegex.FindAllStringSubmatch(src, -1)

   	for _, link := range match {
   		subdomainsfound = append(subdomainsfound, link[1])
	}

	channel <- subdomainsfound
	return
}	

// 
// Query : Queries awesome CertDB service for subdomains
// @param state : current application state, holds all information found
//
func Query(state *helper.State, ch chan helper.Result) {

	var result helper.Result
	result.Subdomains = subdomains

	// Make a http request to CertDB
	resp, err := helper.GetHTTPResponse("https://certdb.com/domain/"+state.Domain, state.Timeout)
	if err != nil {
		result.Error = err
		ch <- result
		return
	}

    // Get the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		result.Error = err
		ch <- result
		return
	}

	src := string(body)

	// Get links for all the SSL Certficates found
	Regex, _ := regexp.Compile("<a href=\"(.*)\" class=\"see-more-link\">See more →</a>")
	match := Regex.FindAllStringSubmatch(src, -1)

	var initialSubs []string
	var subsReturned []string

	channel := make(chan []string, len(match))

   	for _, link := range match {
   		go findSubdomains(link[1], state, channel)
	}

	for i:=0; i < len(match); i++ {
		subsReturned = <-channel

		initialSubs = append(initialSubs, subsReturned...)
	}

	for _, subdomain := range initialSubs {
		if state.Verbose == true {
			if state.Color == true {
				fmt.Printf("\n[%sCERTDB%s] %s", helper.Red, helper.Reset, subdomain)
			} else {
				fmt.Printf("\n[CERTDB] %s", subdomains)
			}
		}

		subdomains = append(subdomains, subdomain)
    }


  	result.Subdomains = subdomains
	result.Error = nil
	ch <-result
}