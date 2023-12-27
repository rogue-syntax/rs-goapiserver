package tls

import (
	//"rs-apiserver.com/global"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"net/http"

	"rs-apiserver.com/global"
)

type HttpClientTY struct {
	httpClient *http.Client
}

var Htc HttpClientTY

func CreateTLSConf() (tls.Config, error) {
	var tfg tls.Config
	rootCertPool := x509.NewCertPool()
	pem, err := ioutil.ReadFile("/var/ssl/kbx-ca-cert.pem")
	if err != nil {
		return tfg, err
	}
	if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
		err := errors.New("Failed to append PEM")
		return tfg, err
	}
	clientCert := make([]tls.Certificate, 0, 1)
	certs, err := tls.LoadX509KeyPair("/var/ssl/kbx-client-cert.pem", "/var/ssl/kbx-client-key.pem")
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
