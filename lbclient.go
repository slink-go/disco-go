package disco_go

import (
	"errors"
	"fmt"
	"github.com/slink-go/disco/common/api"
	"github.com/slink-go/httpclient"
	"github.com/slink-go/logger"
	"net/http"
)

type LoadBalancingHttpClient interface {
	Get(url string) ([]byte, int, error)
	GetWithParams(url string, params map[string]any) ([]byte, int, error)
	//Post(url string, data map[string]any) ([]byte, int, error)
	//PostData(url string, data []byte) ([]byte, int, error)
}

func NewLbClient(reg DiscoRegistry, cln httpclient.Client) LoadBalancingHttpClient {
	return &lbHttpClient{
		registry: reg,
		client:   cln,
	}
}

type lbHttpClient struct {
	registry DiscoRegistry
	client   httpclient.Client
}

func (c *lbHttpClient) geturl(inputUrl string) (url string, err error) {
	proto, svc, path, _, err := parseUrl(inputUrl)
	if err != nil {
		url = inputUrl
		logger.Debug("[disco-go] effective url: %s", url)
		return
	}
	clnt, err := c.registry.Get(svc)
	if err != nil || clnt == nil {
		url = inputUrl
		logger.Debug("[disco-go] effective url: %s", url)
		return
	}
	u, err := clnt.Endpoint(proto)
	if err != nil {
		url = inputUrl
		logger.Debug("[disco-go] effective url: %s", url)
		return
	}
	if proto == api.UnknownEndpoint {
		proto = api.HttpEndpoint
	}
	if proto != api.HttpEndpoint && proto != api.HttpsEndpoint {
		url = ""
		err = fmt.Errorf("invalid protocol: %v", proto)
		logger.Debug("[disco-go] effective url: %s", url)
		return
	}
	// TODO: handle trailing "/" in endpoint or missing "/" in path's beginning
	url = fmt.Sprintf("%s/%s", u, path)
	logger.Debug("[disco-go] effective url: %s", url)
	return
}

func (c *lbHttpClient) Get(url string) ([]byte, int, error) {
	u, err := c.geturl(url)
	if err != nil && u == "" {
		return nil, http.StatusInternalServerError, err
	}
	b, code, err := c.client.Get(u, nil)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	if code >= 400 {
		return nil, code, errors.New(string(b))
	}
	return b, code, nil
}
func (c *lbHttpClient) GetWithParams(inputUrl string, params map[string]any) ([]byte, int, error) {
	p := ""
	for k, v := range params {
		p = fmt.Sprintf("%v=%v&%s", k, v, p)
	}
	return c.Get(fmt.Sprintf("%s/%s", inputUrl, p))
}
