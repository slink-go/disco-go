package disco_go

import (
	"os"
	"strings"
	"time"
)

type DiscoClientConfig struct {
	Login                  string
	Password               string
	Token                  string
	DiscoEndpoints         []string
	ClientName             string
	ClientEndpoints        []string
	ClientMeta             map[string]any
	ClientTimeout          time.Duration
	ClientBreakThreshold   uint
	ClientRetryAttempts    uint
	ClientRetryInterval    time.Duration
	PerformSslVerification bool
	UpdateNotificationChn  chan struct{}
}

func EmptyConfig() *DiscoClientConfig {
	config := &DiscoClientConfig{
		Login:                  "",
		Password:               "",
		Token:                  "",
		DiscoEndpoints:         nil,
		ClientName:             "",
		ClientEndpoints:        nil,
		ClientMeta:             nil,
		ClientTimeout:          0,
		ClientBreakThreshold:   0,
		ClientRetryAttempts:    0,
		ClientRetryInterval:    0,
		PerformSslVerification: true,
		UpdateNotificationChn:  nil,
	}
	return config
}
func DefaultConfig() *DiscoClientConfig {
	ep := getEnvStrings("DISCO_ENDPOINTS", ",")
	t := getEnvString("DISCO_TOKEN")
	l := getEnvString("DISCO_LOGIN")
	p := getEnvString("DISCO_PASSWORD")
	config := &DiscoClientConfig{
		Login:                  l,
		Password:               p,
		Token:                  t,
		DiscoEndpoints:         ep,
		ClientName:             "",
		ClientEndpoints:        nil,
		ClientMeta:             nil,
		ClientTimeout:          0,
		ClientBreakThreshold:   0,
		ClientRetryAttempts:    2,
		ClientRetryInterval:    2 * time.Second,
		PerformSslVerification: true,
		UpdateNotificationChn:  make(chan struct{}, 1),
	}
	return config
}

func (dc *DiscoClientConfig) SkipSslVerify() *DiscoClientConfig {
	dc.PerformSslVerification = false
	return dc
}
func (dc *DiscoClientConfig) WithToken(token string) *DiscoClientConfig {
	dc.Token = token
	return dc
}
func (dc *DiscoClientConfig) WithAuth(login, password string) *DiscoClientConfig {
	dc.Login = login
	dc.Password = password
	return dc
}
func (dc *DiscoClientConfig) WithDisco(endpoints []string) *DiscoClientConfig {
	dc.DiscoEndpoints = endpoints
	return dc
}
func (dc *DiscoClientConfig) WithName(name string) *DiscoClientConfig {
	dc.ClientName = name
	return dc
}
func (dc *DiscoClientConfig) WithEndpoints(endpoints []string) *DiscoClientConfig {
	dc.ClientEndpoints = endpoints
	return dc
}
func (dc *DiscoClientConfig) WithMeta(meta map[string]any) *DiscoClientConfig {
	dc.ClientMeta = meta
	return dc
}
func (dc *DiscoClientConfig) WithTimeout(tm time.Duration) *DiscoClientConfig {
	dc.ClientTimeout = tm
	return dc
}
func (dc *DiscoClientConfig) WithBreaker(threshold uint) *DiscoClientConfig {
	dc.ClientBreakThreshold = threshold
	return dc
}
func (dc *DiscoClientConfig) WithRetry(attempts uint, delay time.Duration) *DiscoClientConfig {
	dc.ClientRetryAttempts = attempts
	dc.ClientRetryInterval = delay
	return dc
}
func (dc *DiscoClientConfig) WithNotificationChn(chn chan struct{}) *DiscoClientConfig {
	dc.UpdateNotificationChn = chn
	return dc
}

func getEnvStrings(key, separator string) []string {
	v := os.Getenv(key)
	if v == "" {
		return nil
	}
	var result []string
	parts := strings.Split(v, separator)
	for _, p := range parts {
		result = append(result, strings.TrimSpace(p))
	}
	return result
}
func getEnvString(key string) string {
	return os.Getenv(key)
}
