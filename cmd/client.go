package cmd

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/nimaaskarian/goje/timer"
	"github.com/r3labs/sse/v2"
	"github.com/spf13/cobra"
)

var (
	outbound_address string
	insecure_tls     bool
)

func init() {
	rootCmd.AddCommand(clientCmd)
	clientCmd.Flags().AddFlagSet(rootFlags())
	clientCmd.Flags().StringVarP(&outbound_address, "outbound-address", "o", "", "address to outbound server to connect to")
	clientCmd.Flags().BoolVar(&insecure_tls, "insecure-tls", false, "don't verify ssl the certificate")
}

var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "subscribe to an outbound server and run an inbound one",
	PreRunE: func(cmd *cobra.Command, args []string) (errout error) {
		return setupConfigForCmd(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if !strings.HasPrefix(outbound_address, "http://") && !strings.HasPrefix(outbound_address, "https://") {
			outbound_address = "http://" + outbound_address
		}
		t := timer.PomodoroTimer{
			Config: &config.Timer,
		}
		httpClient := http.DefaultClient
		if strings.HasPrefix(outbound_address, "https://") {
			var tlsConfig *tls.Config
			if insecure_tls {
				tlsConfig = &tls.Config{
					InsecureSkipVerify: true,
				}
			} else {
				certPool, err := x509.SystemCertPool()
				if err != nil {
					panic(err)
				}
				if caCertPEM, err := os.ReadFile(config.Certfile); err != nil {
					panic(err)
				} else if ok := certPool.AppendCertsFromPEM(caCertPEM); !ok {
					panic("invalid cert in CA PEM")
				}
				clientTLSCert, err := tls.LoadX509KeyPair(config.Certfile, config.Keyfile)
				if err != nil {
					return err
				}
				tlsConfig = &tls.Config{
					RootCAs:      certPool,
					Certificates: []tls.Certificate{clientTLSCert},
				}

			}
			tr := &http.Transport{
				TLSClientConfig: tlsConfig,
			}
			httpClient = &http.Client{
				Transport: tr,
			}
		}
		client := sse.NewClient(outbound_address + "/api/timer/stream")
		client.Connection = httpClient
		config.Timer.OnSet.Append(func(t *timer.PomodoroTimer) {
			content, _ := json.Marshal(t)
			req, err := http.NewRequest("POST", outbound_address+"/api/timer", bytes.NewBuffer(content))
			if err != nil {
				slog.Error("making a request to address failed", "err", err)
				os.Exit(1)
			}
			resp, err := httpClient.Do(req)
			if err != nil {
				slog.Error("sending a request to address failed", "err", err)
			}
			resp.Body.Close()
		})
		go func() {
			err := client.SubscribeRaw(func(msg *sse.Event) {
				json.Unmarshal(msg.Data, &t)
				fmt.Println(string(msg.Data))
				switch string(msg.Event) {
				case "change":
					config.Timer.OnChange.Run(&t)
				case "end":
					config.Timer.OnModeEnd.Run(&t)
				case "start":
					config.Timer.OnModeStart.Run(&t)
				case "pause":
					config.Timer.OnPause.Run(&t)
				}
			})
			if err != nil {
				slog.Error("subscribing to SSE failed", "err", err)
				os.Exit(1)
			}
		}()
		return setupServerAndSignalWatcher(&t)
	},
}
