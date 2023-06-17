package client

import (
	"disco-go/config"
	"disco-go/http"
	"encoding/json"
	"fmt"
	disco "github.com/ws-slink/disco/common/api"
	"github.com/ws-slink/disco/common/util/logger"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type DiscoClient interface {
	Leave() error
	List() []disco.Client
}

func NewDiscoHttpClient(config *config.DiscoClientConfig) (DiscoClient, error) {
	client := &discoClient{
		config:     config,
		httpClient: buildHttpClient(config),
	}
	_, err := client.join(&disco.JoinRequest{
		ServiceId: config.ClientName,
		Endpoints: config.ClientEndpoints,
		Meta:      config.ClientMeta,
	})
	if err != nil {
		return nil, err
	}
	go client.run()
	go client.handleSignals(syscall.SIGINT, syscall.SIGTERM)
	return client, nil
}

func buildHttpClient(config *config.DiscoClientConfig) http.Client {
	client := http.ClientBuilder(config.Token)
	if config.ClientTimeout > 0 {
		client = client.WithTimeout(config.ClientTimeout)
	}
	if config.ClientBreakThreshold > 0 {
		client = client.WithBreaker(config.ClientBreakThreshold)
	}
	if config.ClientRetryAttempts > 0 && config.ClientRetryInterval > 0 {
		client = client.WithRetry(config.ClientRetryAttempts, config.ClientRetryInterval)
	}
	return client
}

type discoClient struct {
	config       *config.DiscoClientConfig
	httpClient   http.Client
	clientId     string
	pingInterval time.Duration
	stopChn      chan struct{}
	joinRequest  *disco.JoinRequest
}

func (dc *discoClient) Leave() error {
	dc.stopChn <- struct{}{}
	close(dc.stopChn)
	return dc.leave()
}
func (dc *discoClient) List() []disco.Client {
	return nil
}

func (dc *discoClient) join(request *disco.JoinRequest) (*disco.JoinResponse, error) {
	var err error
	var buf []byte

	dc.joinRequest = request

	buf, err = json.Marshal(request)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	err = json.Unmarshal(buf, &m)
	if err != nil {
		return nil, err
	}
	for _, ep := range dc.config.DiscoEndpoints {
		buf, _, err = dc.httpClient.Post(fmt.Sprintf("%s/api/join", ep), m)
		if err == nil {
			var jr disco.JoinResponse
			err = json.Unmarshal(buf, &jr)
			if err != nil {
				continue
			}
			dc.clientId = jr.ClientId
			dc.pingInterval = jr.PingInterval.Duration
			logger.Debug("[join] client id: %s", dc.clientId)
			return &jr, nil
		} else {
			logger.Warning("[join] error: %s", err.Error())
		}
	}
	return nil, err
}
func (dc *discoClient) run() {
	dc.stopChn = make(chan struct{})
	ticker := time.NewTicker(dc.pingInterval)
	logger.Debug("[run][%s] started", dc.clientId)
	for {
		select {
		case <-dc.stopChn:
			logger.Debug("[run][%s] finished", dc.clientId)
			ticker.Stop()
			return
		case _ = <-ticker.C:
			pong, err := dc.ping()
			if err != nil {
				//dc.handlePingError()
				continue
			}
			if pong.Response == disco.PongTypeChanged {
				// update clients list
				logger.Debug("[run][%s] update clients list", dc.clientId)
			}
		}
	}
}
func (dc *discoClient) ping() (*disco.Pong, error) {
	var err error
	var code int
	for _, ep := range dc.config.DiscoEndpoints {
		var b []byte
		b, code, err = dc.httpClient.Post(fmt.Sprintf("%s/api/ping?id=%s", ep, dc.clientId), nil)
		if err != nil {
			logger.Warning("[ping][%s] service call error: %s", dc.clientId, err.Error())
			continue
		}
		if code == 404 {
			_ = dc.leave()
			_, err = dc.join(dc.joinRequest)
			if err != nil {
				logger.Warning("[ping] could not reconnect: %s", err.Error())
			}
		}
		var pong disco.Pong
		err = json.Unmarshal(b, &pong)
		if err != nil {
			logger.Warning("[ping] unmarshall error: %s", err.Error())
			continue
		}
		logger.Debug("[ping][%s] pong: %s %s", dc.clientId, pong.Response, pong.Error)
		return &pong, nil
	}
	return nil, err
}
func (dc *discoClient) leave() error {
	var err error
	for _, ep := range dc.config.DiscoEndpoints {
		var b []byte
		b, _, err = dc.httpClient.Post(fmt.Sprintf("%s/api/leave?id=%s", ep, dc.clientId), nil)
		if err != nil {
			logger.Warning("[leave][%s] service call error: %s", dc.clientId, err.Error())
			continue
		}
		logger.Debug("[leave][%s] response: %s", dc.clientId, strings.TrimSpace(string(b)))
		return nil
	}
	return err
}

func (dc *discoClient) handleSignals(signals ...os.Signal) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, signals...)
	done := make(chan bool, 1)
	go func() {
		sig := <-sigs
		logger.Debug("[signal] received %s signal", sig)
		_ = dc.Leave()
		done <- true
	}()
	<-done
}
