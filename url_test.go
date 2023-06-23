package disco_go

import (
	"github.com/slink-go/disco/common/api"
	"testing"
)

const (
	url1 = "http://service/api/v1"
	url2 = "service/api/v1"
	url3 = "service:8081/api/v1"
	url4 = "http://service:8081/api/v1"
	url5 = "http://service.com/api/v1"
	url6 = "http://service.com/api/v1?a=b&c=d%20e"
)

func TestUrlParseA(t *testing.T) {
	pr, sv, pa, ur, err := parseUrl(url1)
	if pr != api.HttpEndpoint {
		t.Errorf("invalid protocol: expected 'http', received '%s'", pr)
	}
	if sv != "service" {
		t.Errorf("invalid service: expected 'service', received '%s'", sv)
	}
	if pa != "api/v1" {
		t.Errorf("invalid path: expected 'api/v1', received '%s'", pa)
	}
	if ur != url1 {
		t.Errorf("invalid url: expected '%s', received '%s'", url1, ur)
	}
	if err != nil {
		t.Errorf("expected no error, %v received", err)
	}
}
func TestUrlParseB(t *testing.T) {
	pr, sv, pa, ur, err := parseUrl(url2)
	if err != nil {
		t.Errorf("expected no error, %v received", err)
	}
	if pr != api.UnknownEndpoint {
		t.Errorf("invalid protocol: expected '', received '%s'", pr)
	}
	if sv != "service" {
		t.Errorf("invalid service: expected 'service', received '%s'", sv)
	}
	if pa != "api/v1" {
		t.Errorf("invalid path: expected 'api/v1', received '%s'", pa)
	}
	if ur != url2 {
		t.Errorf("invalid url: expected '%s', received '%s'", url2, ur)
	}
}
func TestUrlParseC(t *testing.T) {
	pr, sv, pa, ur, err := parseUrl(url3)
	if err != nil {
		t.Errorf("expected no error, %v received", err)
	}
	if pr != api.UnknownEndpoint {
		t.Errorf("invalid protocol: expected '', received '%s'", pr)
	}
	if sv != "service:8081" {
		t.Errorf("invalid service: expected 'service', received '%s'", sv)
	}
	if pa != "api/v1" {
		t.Errorf("invalid path: expected 'api/v1', received '%s'", pa)
	}
	if ur != url3 {
		t.Errorf("invalid url: expected '%s', received '%s'", url3, ur)
	}
}
func TestUrlParseD(t *testing.T) {
	pr, sv, pa, ur, err := parseUrl(url4)
	if err != nil {
		t.Errorf("expected no error, %v received", err)
	}
	if pr != api.HttpEndpoint {
		t.Errorf("invalid protocol: expected '', received '%s'", pr)
	}
	if sv != "service:8081" {
		t.Errorf("invalid service: expected 'service', received '%s'", sv)
	}
	if pa != "api/v1" {
		t.Errorf("invalid path: expected 'api/v1', received '%s'", pa)
	}
	if ur != url4 {
		t.Errorf("invalid url: expected '%s', received '%s'", url4, ur)
	}
}
func TestUrlParseE(t *testing.T) {
	pr, sv, pa, ur, err := parseUrl(url5)
	if err != nil {
		t.Errorf("expected no error, %v received", err)
	}
	if pr != api.HttpEndpoint {
		t.Errorf("invalid protocol: expected '', received '%s'", pr)
	}
	if sv != "service.com" {
		t.Errorf("invalid service: expected 'service', received '%s'", sv)
	}
	if pa != "api/v1" {
		t.Errorf("invalid path: expected 'api/v1', received '%s'", pa)
	}
	if ur != url5 {
		t.Errorf("invalid url: expected '%s', received '%s'", url5, ur)
	}
}
func TestUrlParseF(t *testing.T) {
	pr, sv, pa, ur, err := parseUrl(url6)
	if err != nil {
		t.Errorf("expected no error, %v received", err)
	}
	if pr != api.HttpEndpoint {
		t.Errorf("invalid protocol: expected '', received '%s'", pr)
	}
	if sv != "service.com" {
		t.Errorf("invalid service: expected 'service', received '%s'", sv)
	}
	if pa != "api/v1?a=b&c=d%20e" {
		t.Errorf("invalid path: expected 'api/v1', received '%s'", pa)
	}
	if ur != url6 {
		t.Errorf("invalid url: expected '%s', received '%s'", url6, ur)
	}
}
