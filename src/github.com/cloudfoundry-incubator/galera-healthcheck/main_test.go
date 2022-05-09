package main_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strings"

	"code.cloudfoundry.org/tlsconfig"
	"code.cloudfoundry.org/tlsconfig/certtest"
	"github.com/cloudfoundry-incubator/galera-healthcheck/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Galera Agent", func() {
	var session *gexec.Session
	var serverAuthority *certtest.Authority

	BeforeEach(func() {
		var err error
		serverAuthority, err = certtest.BuildCA("serverCA")
		Expect(err).ToNot(HaveOccurred())

		serverCert, err := serverAuthority.BuildSignedCertificate("serverCert")
		Expect(err).ToNot(HaveOccurred())

		serverCertPEM, serverKeyPEM, err := serverCert.CertificatePEMAndPrivateKey()
		Expect(err).ToNot(HaveOccurred())

		cfg := config.Config{
			DB: config.DBConfig{
				Password: "root-password",
			},
			Monit: config.MonitConfig{
				Host:                          "foo",
				User:                          "foo",
				Port:                          "foo",
				Password:                      "foo",
				MysqlStateFilePath:            "foo",
				ServiceName:                   "foo",
				GaleraInitStatusServerAddress: "foo",
			},
			Host:       "localhost",
			Port:       8080,
			MysqldPath: "mysqld",
			MyCnfPath:  "my.cnf",
			SidecarEndpoint: config.SidecarEndpointConfig{
				Username: "basic-auth-username",
				Password: "basic-auth-password",
			},
			ServerCert: string(serverCertPEM),
			ServerKey:  string(serverKeyPEM),
		}
		b, err := json.Marshal(&cfg)
		Expect(err).NotTo(HaveOccurred())

		cmd := exec.Command(
			binaryPath,
			fmt.Sprintf("-config=%s", string(b)),
		)

		session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		session.Terminate()
	})

	It("Only accepts connections over TLS", func() {
		Eventually(func() error {
			res, err := http.Get("http://127.0.0.1:8080")
			if err != nil {
				return err
			}

			if res.StatusCode == http.StatusOK {
				return nil
			}

			body, _ := ioutil.ReadAll(res.Body)
			trimmedBody := strings.TrimSpace(string(body))
			return fmt.Errorf("received status code: %d, with body: %s", res.StatusCode, trimmedBody)
		}, "10s", "1s").Should(MatchError(`received status code: 400, with body: Client sent an HTTP request to an HTTPS server.`))
	})

	It("Accepts TLS connections", func() {
		serverCertPool, err := serverAuthority.CertPool()
		Expect(err).ToNot(HaveOccurred())

		tlsClientConfig, err := tlsconfig.Build().Client(
			tlsconfig.WithAuthority(serverCertPool),
		)

		httpClient := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlsClientConfig,
			},
		}

		Eventually(func() error {
			res, err := httpClient.Get("https://127.0.0.1:8080/health")
			if err != nil {
				return err
			}

			if res.StatusCode == http.StatusOK {
				return nil
			}

			body, _ := ioutil.ReadAll(res.Body)
			trimmedBody := strings.TrimSpace(string(body))
			return fmt.Errorf("received status code: %d, with body: %s", res.StatusCode, trimmedBody)
		}, "10s", "1s").ShouldNot(HaveOccurred())
	})
})