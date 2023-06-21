package disco_go

import "time"

type DiscoClientConfig struct {
	Token                string
	DiscoEndpoints       []string
	ClientName           string
	ClientEndpoints      []string
	ClientMeta           map[string]any
	ClientTimeout        time.Duration
	ClientBreakThreshold uint
	ClientRetryAttempts  uint
	ClientRetryInterval  time.Duration
}

func DefaultConfig() *DiscoClientConfig {
	return &DiscoClientConfig{
		Token:                "",
		DiscoEndpoints:       nil,
		ClientName:           "",
		ClientEndpoints:      nil,
		ClientMeta:           nil,
		ClientTimeout:        0,
		ClientBreakThreshold: 0,
		ClientRetryAttempts:  0,
		ClientRetryInterval:  0,
	}
}

func (dc *DiscoClientConfig) WithToken(token string) *DiscoClientConfig {
	dc.Token = token
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
