package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/nimaaskarian/goje/timer"
	"github.com/r3labs/sse/v2"
	"github.com/spf13/cobra"
)

var outbound_address string

func init() {
	rootCmd.AddCommand(clientCmd)
	clientCmd.Flags().AddFlagSet(rootFlags())
	clientCmd.Flags().StringVarP(&outbound_address, "outbound-address", "o", "", "address to outbound server to connect to")
}

var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "subscribe to an outbound server and run an inbound one",
	PreRunE: func(cmd *cobra.Command, args []string) (errout error) {
		return setupConfigForCmd(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) (errout error) {
		t := timer.Timer{}
		if err := setupDaemons(&t); err != nil {
			return err
		}
		if !strings.HasSuffix(outbound_address, "http://") && !strings.HasSuffix(outbound_address, "https://") {
			outbound_address = "http://" + outbound_address
		}
		client := sse.NewClient(outbound_address+"/api/timer/stream")
		t.Config.OnSet.Append(func (t *timer.Timer) {
			content, _ := json.Marshal(t)
			req, err := http.NewRequest("POST", outbound_address+"/api/timer", bytes.NewBuffer(content))
			if err != nil {
				errout = err
			}
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				errout = err
			}
			resp.Body.Close()
		})
		go func() {
			err := client.SubscribeRaw(func(msg *sse.Event) {
				json.Unmarshal(msg.Data, &t)
				switch string(msg.Event)  {
					case "change":
						t.Config.OnChange.Run(&t)
					case "end":
						t.Config.OnModeEnd.Run(&t)
					case "start":
						t.Config.OnModeStart.Run(&t)
					case "pause":
						t.Config.OnPause.Run(&t)
				}
			})
			if err != nil {
				errout = err
			}
		}()
		return listenForSignalsForCmdAndTimer(cmd, &t)
	},
}
