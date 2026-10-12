package links

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"golang.org/x/net/publicsuffix"
)

func IsFromSameZone(source, input string) bool {

	base, err := url.Parse(source)
	if err != nil {
		log.Printf("failed to parse source url %s: %v", source, err)
		return false
	}

	tested, err := url.Parse(input)
	if err != nil {
		log.Printf("failed to parse tested url %s: %v", input, err)
		return false
	}

	baseDomain, err := publicsuffix.EffectiveTLDPlusOne(base.Hostname())

	if err != nil {
		log.Printf("failed to get the base domain from %s: %v", base.Hostname(), err)
		return false
	}

	testedDomain, err := publicsuffix.EffectiveTLDPlusOne(tested.Hostname())
	if err != nil {
		log.Printf("failed to get the tested domain from %s: %v", tested.Hostname(), err)
		return false
	}

	return baseDomain == testedDomain

}

func Normalize(input *url.URL) string {

	port := input.Port()
	if len(port) > 0 {
		port = fmt.Sprintf(":%s", port)
	}

	//drop the port value for defaults
	scheme := strings.ToLower(input.Scheme)
	if (scheme == "http" && port == ":80") ||
		(scheme == "https" && port == ":443") {
		port = ""
	}
	if len(scheme) > 0 {
		scheme = fmt.Sprintf("%s://", scheme)
	}

	path := input.EscapedPath()
	if len(path) > 1 && strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}

	query := input.RawQuery
	if len(query) > 0 {
		query = fmt.Sprintf("?%s", query)
	}

	return fmt.Sprintf("%s%s%s%s%s",
		scheme,
		strings.ToLower(input.Hostname()),
		port,
		path,
		query,
	)

}

func Complete(baseUrl, input string) (*url.URL, error) {

	shortie, err := url.Parse(input)
	if err != nil {
		return nil, err
	}

	full, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}

	return full.ResolveReference(shortie), nil

}
