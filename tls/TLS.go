package tls

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"

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
		errx := errors.Wrap(err, err.Error())
		return tfg, errx
	}
	miniopem, err := ioutil.ReadFile(global.EnvVars.MinioSSLCert)
	if err != nil {
		errx := errors.Wrap(err, err.Error())
		return tfg, errx
	}
	if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
		err := errors.New("Failed to append PEM")
		errx := errors.Wrap(err, err.Error())
		return tfg, errx
	}
	if ok := rootCertPool.AppendCertsFromPEM(miniopem); !ok {
		err := errors.New("Failed to append PEM")
		errx := errors.Wrap(err, err.Error())
		return tfg, errx
	}
	clientCert := make([]tls.Certificate, 0, 1)
	certs, err := tls.LoadX509KeyPair(global.EnvVars.SSLCliCert, global.EnvVars.SSLCliKey)
	if err != nil {
		errx := errors.Wrap(err, err.Error())
		return tfg, errx
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
