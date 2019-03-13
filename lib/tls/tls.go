package tls

import (
	"crypto/tls"
	"crypto/x509"
	"log"
	"io/ioutil"
)

// TLSConfig returns a tls.Config, may be nil without error if TLS is not
// configured.
func TLSConfig(TLSCA string, TLSCert string , TLSKey string, Insecure bool) (*tls.Config, error) {
	// TODO: return default tls.Config; plugins should not call if they don't
	// want TLS, this will require using another option to determine.  In the
	// case of an HTTP plugin, you could use `https`.  Other plugins may need
	// the dedicated option `TLSEnable`.

	tlsConfig := &tls.Config{
		InsecureSkipVerify: Insecure,
		Renegotiation:      tls.RenegotiateNever,
	}

	if TLSCA != "" {
		pool, err := makeCertPool([]string{TLSCA})
		if err != nil {
			return nil, err
		}
		tlsConfig.RootCAs = pool
	}

	if TLSCert != "" && TLSKey != "" {
		err := loadCertificate(tlsConfig, TLSCert, TLSKey)
		if err != nil {
			return nil, err
		}
	}

	return tlsConfig, nil
}

func makeCertPool(certFiles []string) (*x509.CertPool, error) {
	pool := x509.NewCertPool()
	for _, certFile := range certFiles {
		pem, err := ioutil.ReadFile(certFile)
		if err != nil {
			log.Fatalln("could not read certificate %q: %v", certFile, err)
		}
		ok := pool.AppendCertsFromPEM(pem)
		if !ok {
			log.Fatalln("could not parse any PEM certificates %q: %v", certFile, err)
		}
	}
	return pool, nil
}

func loadCertificate(config *tls.Config, certFile, keyFile string) error {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatalln("could not load keypair %s:%s: %v", certFile, keyFile, err)
	}

	config.Certificates = []tls.Certificate{cert}
	config.BuildNameToCertificate()
	return nil
}