package links

import (
	"fmt"
	"net/url"
	"strings"
	"unicode"

	"golang.org/x/net/publicsuffix"
)

const (
	separator       = "://"
	schemeByDefault = "https://"
)

func FixURL(input string) (string, error) {

	prefix, _, separatorDetected := strings.Cut(input, separator)
	hasPrefixOfLetters := strings.ContainsFunc(prefix, func(r rune) bool {
		return unicode.IsLetter(r)
	})

	if !separatorDetected || separatorDetected && !hasPrefixOfLetters {

		_, err := url.Parse(schemeByDefault + prefix)
		if err != nil {
			return "", err
		}

		return schemeByDefault + prefix, nil
	}

	if separatorDetected &&
		hasPrefixOfLetters &&
		(prefix == "http" || prefix == "https") {

		_, err := url.Parse(input)
		if err != nil {
			return "", err
		}
		return input, nil
	}

	return "", fmt.Errorf("failed to fix the url")

}

func IsFromSameZone(source, input string) (bool, error) {

	base, err := url.Parse(source)
	if err != nil {
		return false, fmt.Errorf("parse source url %s: %w", source, err)
	}

	tested, err := url.Parse(input)
	if err != nil {
		return false, fmt.Errorf("parse tested url %s: %w", input, err)
	}

	baseDomain, err := publicsuffix.EffectiveTLDPlusOne(base.Hostname())

	if err != nil {
		return false, fmt.Errorf("get the base domain from %s: %w", base.Hostname(), err)
	}

	testedDomain, err := publicsuffix.EffectiveTLDPlusOne(tested.Hostname())
	if err != nil {
		return false, fmt.Errorf("get the tested domain from %s: %w", tested.Hostname(), err)
	}

	return baseDomain == testedDomain, nil

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
