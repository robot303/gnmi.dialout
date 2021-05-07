package dialout

import (
	"log"
	"testing"
	"time"

	"github.com/openconfig/gnmi/proto/gnmi"
)

func TestTLS(t *testing.T) {
	// Logging

	Print = log.Print
	Printf = log.Printf
	address := "localhost:8088"

	type server struct {
		insecure   bool
		skipverify bool
		cafile     string
		serverCert string
		serverKey  string
		username   string
		password   string
	}
	type client struct {
		serverName string
		insecure   bool
		skipverify bool
		cafile     string
		clientCert string
		clientKey  string
		username   string
		password   string
	}
	tests := []struct {
		name    string
		server  server
		client  client
		wantErr bool
	}{
		{
			name: "tls setup - set all",
			server: server{
				insecure:   false,
				skipverify: false,
				cafile:     "tls/ca.crt",
				serverCert: "tls/server.crt",
				serverKey:  "tls/server.key",
				username:   "myaccount",
				password:   "mypassword",
			},
			client: client{
				serverName: "hfrnet.com", // server name must be server's SAN (Subject Alternative Name).
				insecure:   false,
				skipverify: false,
				cafile:     "tls/ca.crt",
				clientCert: "tls/client.crt",
				clientKey:  "tls/client.key",
				username:   "myaccount",
				password:   "mypassword",
			},
		},
		{
			name: "tls setup - skip-verify", // in skip-verify mode, server.crt, server.key are only required.
			server: server{
				insecure:   false,
				skipverify: true,
				serverCert: "tls/server.crt",
				serverKey:  "tls/server.key",
				username:   "myaccount",
				password:   "mypassword",
			},
			client: client{
				// serverName: "hfrnet.com",
				insecure:   false,
				skipverify: true,
				username:   "myaccount",
				password:   "mypassword",
			},
		},
		{
			name: "tls setup - insecure",
			server: server{
				insecure:   true,
				skipverify: false,
				// cafile:     "tls/ca.crt",
				// serverCert: "tls/server.crt",
				// serverKey:  "tls/server.key",
				username: "myaccount",
				password: "mypassword",
			},
			client: client{
				// serverName: "hfrnet.com",
				insecure:   true,
				skipverify: false,
				// cafile:     "tls/ca.crt",
				// clientCert: "tls/client.crt",
				// clientKey:  "tls/client.key",
				username: "myaccount",
				password: "mypassword",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, err := NewGNMIDialoutServer(
				address, tt.server.insecure, tt.server.skipverify,
				tt.server.cafile, tt.server.serverCert, tt.server.serverKey,
				tt.server.username, tt.server.password)
			if err != nil {
				t.Error(err)
				return
			}
			go server.Serve()
			defer func() {
				server.Close()
				time.Sleep(time.Millisecond * 10)
			}()
			client, err := NewGNMIDialOutClient(
				tt.client.serverName, address, tt.client.insecure, tt.client.skipverify,
				tt.client.cafile, tt.client.clientCert, tt.client.clientKey,
				tt.client.username, tt.client.password, true)
			if err != nil {
				t.Error(err)
				return
			}
			defer func() {
				client.Close()
				time.Sleep(time.Millisecond * 10)
			}()
			if err := client.Send(
				[]*gnmi.SubscribeResponse{
					&gnmi.SubscribeResponse{
						Response: &gnmi.SubscribeResponse_SyncResponse{
							SyncResponse: true,
						},
					},
				},
			); err != nil {
				t.Error(err)
				return
			}
			time.Sleep(time.Millisecond * 10)
		})
	}
}

func TestGNMIDialOut(t *testing.T) {
	// Logging
	// Print = log.Print
	// Printf = log.Printf

	address := "localhost:8088"
	insecure := true

	server, err := NewGNMIDialoutServer(address, insecure, false, "", "", "", "", "")
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		server.Close()
		time.Sleep(time.Millisecond * 10)
	}()
	go server.Serve()
	client, err := NewGNMIDialOutClient("", address, insecure, false, "", "", "", "", "", true)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		client.Close()
		time.Sleep(time.Millisecond * 10)
	}()
	if err := client.Send(
		[]*gnmi.SubscribeResponse{
			&gnmi.SubscribeResponse{
				Response: &gnmi.SubscribeResponse_SyncResponse{
					SyncResponse: true,
				},
			},
			&gnmi.SubscribeResponse{
				Response: &gnmi.SubscribeResponse_Update{
					Update: &gnmi.Notification{
						Timestamp: 0,
						Prefix: &gnmi.Path{
							Elem: []*gnmi.PathElem{
								&gnmi.PathElem{
									Name: "interfaces",
								},
								&gnmi.PathElem{
									Name: "interface",
									Key: map[string]string{
										"name": "1/1",
									},
								},
							},
						},
						Alias: "#1/1",
					},
				},
			},
			&gnmi.SubscribeResponse{
				Response: &gnmi.SubscribeResponse_Update{
					Update: &gnmi.Notification{
						Timestamp: 0,
						Prefix: &gnmi.Path{
							Elem: []*gnmi.PathElem{
								&gnmi.PathElem{
									Name: "#1/1",
								},
							},
						},
						Update: []*gnmi.Update{
							&gnmi.Update{
								Path: &gnmi.Path{
									Elem: []*gnmi.PathElem{
										&gnmi.PathElem{
											Name: "state",
										},
										&gnmi.PathElem{
											Name: "counters",
										},
										&gnmi.PathElem{
											Name: "in-pkts",
										},
									},
								},
								Val: &gnmi.TypedValue{
									Value: &gnmi.TypedValue_UintVal{
										UintVal: 100,
									},
								},
							},
						},
					},
				},
			},
		},
	); err != nil {
		t.Error(err)
		return
	}
	time.Sleep(time.Millisecond * 10)
}
