package disco_go

import (
	"github.com/slink-go/disco/common/api"
	"regexp"
	"strings"
)

var regex *regexp.Regexp

func init() {
	regex = regexp.MustCompile(urlRegexp)
}

const urlRegexp = `(?P<PROTO>[a-z]*:\/\/)?(?P<SERVICE>[a-zA-Z0-9-_.:]+)\/(?P<PATH>.*)`

func parseUrl(inputUrl string) (proto api.EndpointType, service string, path string, url string, err error) {
	match := regex.FindStringSubmatch(inputUrl)
	pmap := make(map[string]string)
	for i, name := range regex.SubexpNames() {
		if i > 0 && i < len(match) {
			pmap[name] = match[i]
		}
	}
	proto = api.UnknownEndpoint
	switch strings.ToLower(strings.ReplaceAll(pmap["PROTO"], "://", "")) {
	case "http":
		proto = api.HttpEndpoint
	case "https":
		proto = api.HttpsEndpoint
	case "grpc":
		proto = api.GrpcEndpoint
	}
	service = pmap["SERVICE"]
	path = pmap["PATH"]
	url = inputUrl
	err = nil
	return
}
