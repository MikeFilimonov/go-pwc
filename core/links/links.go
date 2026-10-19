package links

import (
	"errors"
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

var (
	ErrEmptyHost         = errors.New("please, provide a non-empty url")
	ErrUnsupportedHost   = errors.New("localhost and ip hosts will be supported later")
	ErrUnsupportedScheme = errors.New("unsupported scheme")
)

func FixURL(input string) (string, error) {

	result := input

	prefix, _, separatorDetected := strings.Cut(input, separator)
	hasPrefixOfLetters := !strings.ContainsFunc(prefix, func(r rune) bool {
		return !unicode.IsLetter(r)
	})

	hasScheme := separatorDetected && hasPrefixOfLetters

	if hasScheme {
		prefix = strings.ToLower(prefix)
	}

	isValidPrefix := prefix == "http" || prefix == "https"

	if hasScheme && !isValidPrefix {
		return "", ErrUnsupportedScheme
	}

	if !separatorDetected || !hasScheme {
		result = schemeByDefault + input
	}

	parsedUrl, err := url.Parse(result)
	if err != nil {
		return "", err
	}

	if len(parsedUrl.Hostname()) == 0 {
		return "", ErrEmptyHost
	}

	_, err = publicsuffix.EffectiveTLDPlusOne(parsedUrl.Hostname())

	if err != nil {
		return "", ErrUnsupportedHost
	}

	return result, nil

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
