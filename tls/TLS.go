package tls

import (
	//"github.com/rogue-syntax/rs-goapiserver/global"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/rogue-syntax/rs-goapiserver/global"
)

type HttpClientTY struct {
	httpClient *http.Client
}

var Htc HttpClientTY

func CreateTLSConf() (tls.Config, error) {
	var tfg tls.Config
	rootCertPool := x509.NewCertPool()
	pem, err := ioutil.ReadFile(global.EnvVars.SSLCaCert)
	if err != nil {
		return tfg, err
	}
	if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
		err := errors.New("Failed to append PEM")
		return tfg, err
	}
	clientCert := make([]tls.Certificate, 0, 1)
	certs, err := tls.LoadX509KeyPair(global.EnvVars.SSLCliCert, global.EnvVars.SSLCliKey)
	if err != nil {
		return tfg, err
	}

	Htc.httpClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:            rootCertPool,
				InsecureSkipVerify: true,
			},
		},
	}

	clientCert = append(clientCert, certs)
	return tls.Config{
		RootCAs:            rootCertPool,
		Certificates:       clientCert,
		ServerName:         global.EnvVars.Dbserver,
		InsecureSkipVerify: true, // needed for self signed certs
	}, nil

}
