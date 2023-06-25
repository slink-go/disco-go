package disco_go

import (
	"encoding/json"
	"errors"
	"fmt"
	disco "github.com/slink-go/disco/common/api"
	http "github.com/slink-go/httpclient"
	"github.com/slink-go/logger"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

func NewDiscoHttpClient(config *DiscoClientConfig) (DiscoClient, error) {
	dClient := &discoClientImpl{
		config:     config,
		httpClient: buildHttpClient(config),
		registry:   NewRegistry(),
	}
	_, err := dClient.join(&disco.JoinRequest{
		ServiceId: config.ClientName,
		Endpoints: config.ClientEndpoints,
		Meta:      config.ClientMeta,
	})
	if err != nil {
		return nil, err
	}
	if err = dClient.sync(); err != nil {
		logger.Warning("sync error: %s", err.Error())
	}
	go dClient.run()
	go dClient.handleSignals(syscall.SIGINT, syscall.SIGTERM)
	return dClient, nil
}

func buildHttpClient(config *DiscoClientConfig) http.Client {
	clnt := http.NewTokenAuthClient(config.Token)
	if config.ClientTimeout > 0 {
		clnt = clnt.WithTimeout(config.ClientTimeout)
	}
	if config.ClientBreakThreshold > 0 {
		clnt = clnt.WithBreaker(config.ClientBreakThreshold)
	}
	if config.ClientRetryAttempts > 0 && config.ClientRetryInterval > 0 {
		clnt = clnt.WithRetry(config.ClientRetryAttempts, config.ClientRetryInterval)
	}
	return clnt
}

type discoClientImpl struct {
	config       *DiscoClientConfig
	httpClient   http.Client
	clientId     string
	pingInterval time.Duration
	stopChn      chan struct{}
	joinRequest  *disco.JoinRequest
	registry     DiscoRegistry
	mutex        sync.Mutex
}

func (dc *discoClientImpl) Leave() error {
	dc.stopChn <- struct{}{}
	close(dc.stopChn)
	return dc.leave()
}
func (dc *discoClientImpl) Registry() DiscoRegistry {
	return dc.registry
}

func (dc *discoClientImpl) join(request *disco.JoinRequest) (*disco.JoinResponse, error) {
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
	res, _, err := dc.oneOf(
		"join",
		dc.httpClient.Post,
		joinUrlTemplate,
		m,
	)
	if err != nil {
		//logger.Warning("[join] error: %s", err.Error())
		return nil, err
	}

	var jr disco.JoinResponse
	err = json.Unmarshal(res, &jr)
	if err != nil {
		return nil, err
	}
	dc.clientId = jr.ClientId
	dc.pingInterval = jr.PingInterval.Duration
	logger.Debug("[join] client id: %s", dc.clientId)
	return &jr, nil
}
func (dc *discoClientImpl) run() {
	dc.stopChn = make(chan struct{})
	pingTicker := time.NewTicker(dc.pingInterval)
	syncTicker := time.NewTicker(time.Duration(10) * dc.pingInterval)
	logger.Debug("[run][%s] started", dc.clientId)
	for {
		select {
		case <-dc.stopChn:
			logger.Debug("[run][%s] finished", dc.clientId)
			pingTicker.Stop()
			syncTicker.Stop()
			return
		case _ = <-pingTicker.C:
			pong, err := dc.ping()
			if err != nil {
				logger.Warning("ping error: %s", err.Error())
				continue
			}
			if pong.Response == disco.PongTypeChanged {
				logger.Debug("[run][%s] update clients list", dc.clientId)
				// random jitter
				time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
				err = dc.sync()
				if err != nil {
					logger.Warning("sync error: %s", err.Error())
				}
			}
		case _ = <-syncTicker.C:
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			err := dc.sync()
			if err != nil {
				logger.Warning("sync error: %s", err.Error())
			}
		}
	}
}
func (dc *discoClientImpl) ping() (*disco.Pong, error) {
	res, code, err := dc.oneOf(
		"ping",
		dc.httpClient.Post,
		fmt.Sprintf(pingUrlTemplate, "%s", dc.clientId),
		nil,
	)
	if code == 404 {
		go func() {
			_ = dc.leave()
		}()
		_, err = dc.join(dc.joinRequest)
		if err != nil {
			logger.Warning("[ping] could not reconnect: %s", err.Error())
		}
		return &disco.Pong{}, nil
	}
	var pong disco.Pong
	err = json.Unmarshal(res, &pong)
	if err != nil {
		logger.Warning("[ping] unmarshall error: %s", err.Error())
		return nil, err
	}
	logger.Debug("[ping][%s] pong: %s %s", dc.clientId, pong.Response, pong.Error)
	return &pong, nil
}
func (dc *discoClientImpl) leave() error {
	_, _, err := dc.oneOf(
		"leave",
		dc.httpClient.Post,
		fmt.Sprintf(leaveUrlTemplate, "%s", dc.clientId),
		nil,
	)
	return err
}
func (dc *discoClientImpl) sync() error {
	dc.mutex.Lock()
	defer dc.mutex.Unlock()

	res, _, err := dc.oneOf(
		"list",
		dc.httpClient.Get,
		listUrlTemplate,
		nil,
	)
	if err != nil {
		return err
	}

	var clients []client
	err = json.Unmarshal(res, &clients)
	if err != nil {
		return err
	}

	var result []Client
	for i := 0; i < len(clients); i++ {
		result = append(result, &clients[i])
	}
	dc.Registry().Sync(result)

	return nil
}

func (dc *discoClientImpl) handleSignals(signals ...os.Signal) {
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

// oneOf - calls first available endpoint from all endpoints configured for service
func (dc *discoClientImpl) oneOf(action string, call serviceCall, urlTemplate string, args map[string]any) ([]byte, int, error) {
	var err error
	var code int
	for _, ep := range dc.config.DiscoEndpoints {
		var b []byte
		b, code, err = call(fmt.Sprintf(urlTemplate, ep), args)
		if err != nil {
			if errors.Is(err, &http.HttpError{}) {
				logger.Warning("[%s][%s] service call error (%d): %s", action, dc.clientId, code, err.Error())
			} else {
				logger.Warning("[%s][%s] service call error: %s", action, dc.clientId, err.Error())
			}
			continue
		}
		logger.Debug("[%s][%s] response: %d, %s", action, dc.clientId, code, strings.TrimSpace(string(b)))
		return b, code, err
	}
	return nil, code, err
}
