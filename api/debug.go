package api

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"os"
)

type debugInfo struct {
	Server *serverInfo `json:"server"`
	IP     *ipInfo     `json:"ip"`
	HTTP   *httpInfo   `json:"http"`
	TLS    *tlsInfo    `json:"tls"`
}

type serverInfo struct {
	Hostname string `json:"hostname"`
}

type ipInfo struct {
	RemoteAddr string `json:"remote_addr"`
}

type httpInfo struct {
	Protocol string      `json:"protocol"`
	Method   string      `json:"method"`
	Header   http.Header `json:"header"`
	URL      string      `json:"url"`
}

type tlsInfo struct {
	Version            uint16 `json:"version"`
	DidResume          bool   `json:"did_resume"`
	CipherSuite        uint16 `json:"cipher_suite"`
	NegotiatedProtocol string `json:"negotiated_protocol"`
	ServerName         string `json:"server_name"`
}

func (api *API) handleDebug(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "\t")

	if r.TLS == nil {
		r.TLS = &tls.ConnectionState{}
	}

	enc.Encode(&debugInfo{ //nolint:errcheck,errchkjson
		Server: buildServerInfo(),
		IP:     buildIPInfo(r),
		HTTP:   buildHTTPInfo(r),
		TLS:    buildTLSInfo(r),
	})
}

func buildServerInfo() *serverInfo {
	hostname, _ := os.Hostname()

	return &serverInfo{
		Hostname: hostname,
	}
}

func buildIPInfo(r *http.Request) *ipInfo {
	return &ipInfo{
		RemoteAddr: r.RemoteAddr,
	}
}

func buildHTTPInfo(r *http.Request) *httpInfo {
	return &httpInfo{
		Protocol: r.Proto,
		Method:   r.Method,
		Header:   r.Header,
		URL:      r.URL.String(),
	}
}

func buildTLSInfo(r *http.Request) *tlsInfo {
	return &tlsInfo{
		Version:            r.TLS.Version,
		DidResume:          r.TLS.DidResume,
		CipherSuite:        r.TLS.CipherSuite,
		NegotiatedProtocol: r.TLS.NegotiatedProtocol,
		ServerName:         r.TLS.ServerName,
	}
}
