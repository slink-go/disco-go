package disco_go

import (
	"encoding/json"
	"errors"
	"fmt"
	disco "github.com/slink-go/disco/common/api"
	httpc "github.com/slink-go/httpclient"
	"github.com/slink-go/logging"
	"math/rand"
	"net/http"
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
		logger:     logging.GetLogger("disco-go"),
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
		dClient.logger.Warning("[init] sync error: %s", err.Error())
	}
	go dClient.run()
	go dClient.handleSignals(syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	return dClient, nil
}

func NewDiscoHttpClientPanicOnAuthError(config *DiscoClientConfig) DiscoClient {
	for {
		clnt, err := NewDiscoHttpClient(config)
		if err != nil {
			switch err.(type) {
			case *httpc.HttpError:
				e := err.(*httpc.HttpError)
				if e.Code() == http.StatusUnauthorized || e.Code() == http.StatusForbidden {
					panic(e)
				}
			}
			time.Sleep(time.Second)
			continue
		}
		return clnt
	}
}

func buildHttpClient(config *DiscoClientConfig) httpc.Client {
	clnt := httpc.New()
	if !config.PerformSslVerification {
		clnt = clnt.SkipTlsVerify()
	}
	if config.Token != "" {
		clnt = clnt.WithBearerAuth(config.Token)
	} else if config.Login != "" && config.Password != "" {
		clnt = clnt.WithBasicAuth(config.Login, config.Password)
	} else {
		clnt = clnt.WithNoAuth()
	}
	if config.ClientTimeout > 0 {
		clnt = clnt.WithTimeout(config.ClientTimeout)
	}
	if config.ClientBreakThreshold > 0 {
		clnt = clnt.WithBreaker(
			config.ClientBreakThreshold,
			time.Second,
			15*time.Second,
		)
	}
	if config.ClientRetryAttempts > 0 && config.ClientRetryInterval > 0 {
		clnt = clnt.WithRetry(
			config.ClientRetryAttempts,
			config.ClientRetryInterval,
		)
	}
	return clnt
}

type discoClientImpl struct {
	config       *DiscoClientConfig
	logger       logging.Logger
	httpClient   httpc.Client
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

	dc.logger.Debug("[join] try join '%s'", request.ServiceId)
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
		dc.logger.Warning("[join] error: %s", err.Error())
		return nil, err
	}

	var jr disco.JoinResponse
	err = json.Unmarshal(res, &jr)
	if err != nil {
		return nil, err
	}
	dc.clientId = jr.ClientId
	dc.pingInterval = jr.PingInterval.Duration
	dc.logger.Debug("[join][%s] connected to disco server", dc.clientId)
	return &jr, nil
}
func (dc *discoClientImpl) run() {
	dc.stopChn = make(chan struct{})
	pingTicker := time.NewTicker(dc.pingInterval)
	dc.logger.Debug("[run][%s] started", dc.clientId)
	for {
		select {
		case <-dc.stopChn:
			dc.logger.Debug("[run][%s] finished", dc.clientId)
			pingTicker.Stop()
			return
		case _ = <-pingTicker.C:
			pong, err := dc.ping()
			dc.logger.Trace("ping response: %s", pong)
			if err != nil {
				dc.logger.Warning("ping error: %s", err.Error())
				continue
			}
			if pong.Response == disco.PongTypeChanged {
				dc.logger.Debug("[run][%s] update clients list", dc.clientId)
				time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond) // add random jitter
				err = dc.sync()
				if err != nil {
					dc.logger.Warning("sync error: %s", err.Error())
				}
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

	//logger.Warning("ping: code=%d, err=%v", code, err)
	if code == 404 || code == httpc.HttpClientRetriesExhaustedError {
		go func() {
			_ = dc.leave()
		}()
		_, err = dc.join(dc.joinRequest)
		if err != nil {
			dc.logger.Warning("[ping] could not reconnect: %s", err.Error())
			return &disco.Pong{}, nil
		}
		return &disco.Pong{Response: disco.PongTypeChanged}, nil
	}

	if err != nil {
		if strings.Contains(err.Error(), "connection refused") {
			dc.logger.Warning("[ping] ping error; reset clients cache: %s", err.Error())
			dc.registry.Reset()
			return &disco.Pong{}, nil
		}
		return &disco.Pong{}, err
	}

	var pong disco.Pong
	err = json.Unmarshal(res, &pong)
	if err != nil {
		dc.logger.Warning("[ping] unmarshall error: %s", err.Error())
		return nil, err
	}
	//logger.Debug("[ping][%s] pong: %s %s", dc.clientId, pong.Response, pong.Error)
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
	sigs := make(chan os.Signal)
	signal.Notify(sigs, signals...)
	for {
		sig := <-sigs
		switch sig {
		case syscall.SIGINT:
			fallthrough
		case syscall.SIGKILL:
			fallthrough
		case syscall.SIGTERM:
			dc.logger.Debug("[signal] received %s signal", sig)
			dc.logger.Trace("[signal] leave")
			_ = dc.Leave()
			//time.Sleep(time.Millisecond * 100)
			dc.logger.Trace("[signal] done")
		}
	}
}

// oneOf - calls first available endpoint from all endpoints configured for service
func (dc *discoClientImpl) oneOf(action string, call serviceCall, urlTemplate string, args map[string]any) ([]byte, int, error) {
	var err error
	var code int
	for _, ep := range dc.config.DiscoEndpoints {
		var b []byte
		b, _, code, err = call(fmt.Sprintf(urlTemplate, ep), args, nil)
		err = dc.processError(err)
		if err != nil {
			if errors.Is(err, &httpc.HttpError{}) ||
				errors.Is(err, &httpc.ConnectionRefusedError{}) ||
				errors.Is(err, &httpc.ServiceUnreachableError{}) {
				dc.logger.Warning("[%s][%s] service call error: message='%s', code=%d", action, dc.clientId, strings.TrimSpace(err.Error()), code)
			} else {
				dc.logger.Warning("[%s][%s] service call error: %s", action, dc.clientId, strings.TrimSpace(err.Error()))
			}
			continue
		}
		//logger.Debug("[%s][%s] response: %d, %s", action, dc.clientId, code, strings.TrimSpace(string(b)))
		return b, code, err
	}
	return nil, code, err
}

func (dc *discoClientImpl) processError(err error) error {
	//if err == nil {
	//	return nil
	//}
	//if strings.Contains(err.Error(), "service unreachable")
	return err
}
