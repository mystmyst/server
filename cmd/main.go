// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 mochi-mqtt, mochi-co
// SPDX-FileContributor: mochi-co

package main

import (
	"crypto/tls"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
)

var (
	wsCetificate = []byte(`-----BEGIN CERTIFICATE-----
	MIIE/DCCA+SgAwIBAgISAwyK1gZZIZ0A+m1klMrhF+gKMA0GCSqGSIb3DQEBCwUA
	MDIxCzAJBgNVBAYTAlVTMRYwFAYDVQQKEw1MZXQncyBFbmNyeXB0MQswCQYDVQQD
	EwJSMzAeFw0yNDAzMjAwMDI1NDdaFw0yNDA2MTgwMDI1NDZaMBkxFzAVBgNVBAMT
	DndzLm9oa2kuY29tLmJyMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA
	vbrGO1UqAWwhC+zIKnY6fLCSPQy7wesPkCjC/n5u3UJ8vb2vKgHGpN2ytCqx3ctp
	WYT+jSZ2LI0quD8kgRtNQ7HtK4KmF/x6yG8NmuYq9fapqSwjH+1y505fX4p9wQZS
	mLyaam9Cp3p11+KI16jnOtCQX8R1ZFUqoRjt1nM71HBjJPcwWENd5BsDfg/w4FB0
	Ek8Fp18OKl9B2GlF1slY6TXh+cQCaENIFlKTgMqvMFVvffLTNHSbG4jBQ/mioxy/
	9KE0ECMBj3UiGx46qDylFC0WAFDXW9FrnqO8p6MoE/p/ATrXDRRHMQaXCrA+Xoql
	RljPKRrHU1b13BpYAsB9xwIDAQABo4ICIzCCAh8wDgYDVR0PAQH/BAQDAgWgMB0G
	A1UdJQQWMBQGCCsGAQUFBwMBBggrBgEFBQcDAjAMBgNVHRMBAf8EAjAAMB0GA1Ud
	DgQWBBRUYaPK7Yg4UNk6IYe1y4t2FCyjgjAfBgNVHSMEGDAWgBQULrMXt1hWy65Q
	CUDmH6+dixTCxjBVBggrBgEFBQcBAQRJMEcwIQYIKwYBBQUHMAGGFWh0dHA6Ly9y
	My5vLmxlbmNyLm9yZzAiBggrBgEFBQcwAoYWaHR0cDovL3IzLmkubGVuY3Iub3Jn
	LzAtBgNVHREEJjAkgg53cy5vaGtpLmNvbS5icoISd3d3LndzLm9oa2kuY29tLmJy
	MBMGA1UdIAQMMAowCAYGZ4EMAQIBMIIBAwYKKwYBBAHWeQIEAgSB9ASB8QDvAHUA
	SLDja9qmRzQP5WoC+p0w6xxSActW3SyB2bu/qznYhHMAAAGOWXbixQAABAMARjBE
	AiBkN4Ah9R2dA5ufNLdDpKBv+/sn1l982UBViJTHci7zLwIgVj2C67AzUdKVUf8h
	1dC4IQWyEAOVx8/rO+h82G+GWPIAdgDuzdBk1dsazsVct520zROiModGfLzs3sNR
	SFlGcR+1mwAAAY5ZduLYAAAEAwBHMEUCIQC1mpES9XFDFNYGsEhXelessPRi1GD8
	pSTEA12LHoLKpgIgXv1gbtbDkH0x2eUMbjwmGdtedOX1aw9wIdlVyBplorAwDQYJ
	KoZIhvcNAQELBQADggEBAADinsMCmQAzRMJzO5OaajqrucnksBXs2kRyD3CTeC40
	2bz7C7XwLza2Wo0T/SeJQlriLPzapyLazTvZ46jY9PrPIBg9YdAhQ6NsIlpwQu/a
	mVvVfgTinPxjzfTZeNSPgjaqupXRz6aBmExqDPUebMwx0JXvuU9RbBPQL3h2WOrg
	G00ZE5Qxa2NoG2BfmxvdK6+U7FPa/Kq1Y7O261mg0pbVZ9EnB2aq5ok2064Jj6dL
	hOKgcG2ntdnZ6wpID4X6TPVVizafTAElQaCyt6C1tzO/wng682C7873SZrq+WvnA
	jhkp1z5vmKEg7MwwmQkHAnZTtL2K47kh6tpoxTd+pGo=
	-----END CERTIFICATE-----`)

	wsPrivateKey = []byte(`-----BEGIN RSA PRIVATE KEY-----
	MIIEpQIBAAKCAQEAvbrGO1UqAWwhC+zIKnY6fLCSPQy7wesPkCjC/n5u3UJ8vb2v
	KgHGpN2ytCqx3ctpWYT+jSZ2LI0quD8kgRtNQ7HtK4KmF/x6yG8NmuYq9fapqSwj
	H+1y505fX4p9wQZSmLyaam9Cp3p11+KI16jnOtCQX8R1ZFUqoRjt1nM71HBjJPcw
	WENd5BsDfg/w4FB0Ek8Fp18OKl9B2GlF1slY6TXh+cQCaENIFlKTgMqvMFVvffLT
	NHSbG4jBQ/mioxy/9KE0ECMBj3UiGx46qDylFC0WAFDXW9FrnqO8p6MoE/p/ATrX
	DRRHMQaXCrA+XoqlRljPKRrHU1b13BpYAsB9xwIDAQABAoIBAQCOjveB/3THitKt
	3iVs2lcJ97Z6DsZJZ/DStf4GMUPmFp4aB5vFKX5zxG0ROP9akwu+itKlhl/HC+8s
	b61jIPuGQPve9JUOctRjJCaJ3CYtmEBU7+gYhlcO+/FnnWzuC20mfJheHulrY/WF
	2B5QRQYxSCMjAj/euquETnHu77jl3pqqyR9LqrUvInQoK7cds1A9SOQ/Gm+KpKds
	8nHbWo6z6UBz4ujUVl0BD+Wwkhnxp/x8WzJkk1vOma8eGc+zbnthbB6hqyOTu8tA
	lJzemddL2oHEiYZOEgomT1MGLr9IHZDgm1Hsi0Y3KiBZCc/USM41+99AkHhW+Ef2
	+zuCNZFpAoGBAPmnwVgJkdlaxlzuMG4B/8BuUrmq0CsgNUvfhjG9w/9a7XO1NGLW
	9n0Idndbc1U/D5aM4ngmVEQ2d4IWf9aRHAuDtJD1nllKPBaJlfmHpvEpqVLsd374
	sm5T3jCs4I5/48xplOmRdGYix9FB2KSTMzBn2lOjOq4/LX00w6/7KSWTAoGBAMKN
	JT1nD0xJiLQrQgkfdpXEA5vQEGYhQsDL9IwUbE9ZKasHlAsvTFF9k0r9LqnftkN4
	yHQiiv6AWak7fz8V1U/ssU8mlz3NMBvSqZC9O9f51MPba7M7RWrJE7fRTo1LbVbe
	4+YR/FmMjR6iguN0TTJnnA0xlcBxlbLHcIMfKWd9AoGBAOUFF0CDxt/1ffLSLms8
	Ojl0+z6Hi9+D9GBd9OS8iIhACYQTvrLNL+ETWlmz8uFIsCwToc1GnBbXQFp9+VgE
	Vg3aDFLOfyy6BNVH8eSupF6nMUV410YLLuQ226UbcgRHHdnvIUQCwxzO2y8DkJGo
	11SYcJg5LSOboUcymDFf3icxAoGBALlrhmuQFs9xYf29ILHLL90rNPlCgu6jkphn
	ikobiOLTKthbX6iNSqJ8GW6mANxcX7zMl9e/uFM5Brs4/lyktWn4P0EdmZWIQuqx
	i3RsNmXwMOz96hanTdCplcZikQgvNCVQR0pWJ/k88J6a6j5X8N8ySlN0x7HjT3ZV
	iJEfmPmNAoGAWcl8JHxz+TPXmyzWuCIJRe96HgoccSNfzQkN6Czu5UKJn4d8WRl9
	ERU6oRCWIzmUFRKhtxVgGUXS2aoz15i7EW5OBuHTipZM+4yCyQTcSJrCt0KOEllJ
	7agFl7hNiEec9Py+sIFxiYwEZVccvEDFFT8tbv2QN4+VUCq4U//HHJY=
	-----END RSA PRIVATE KEY-----`)
)

func main() {

	wsCert, err := tls.X509KeyPair(wsCetificate, wsPrivateKey)
	if err != nil {
		log.Fatal(err)
	}

	wsTlsConfig := &tls.Config{
		Certificates: []tls.Certificate{wsCert},
	}

	tcpAddr := flag.String("tcp", ":1883", "network address for TCP listener")
	wsAddr := flag.String("ws", ":1882", "network address for Websocket listener")
	infoAddr := flag.String("info", ":8080", "network address for web info dashboard listener")
	flag.Parse()

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		done <- true
	}()

	data, err := os.ReadFile("env/auth.yaml")
	if err != nil {
		log.Fatal(err)
	}

	server := mqtt.New(nil)
	//_ = server.AddHook(new(auth.AllowHook), nil)
	err = server.AddHook(new(auth.Hook), &auth.Options{
		Data: data, // build ledger from byte slice, yaml or json
	})
	if err != nil {
		log.Fatal(err)
	}

	tcp := listeners.NewTCP(listeners.Config{
		ID:      "t1",
		Address: *tcpAddr,
	})
	err = server.AddListener(tcp)
	if err != nil {
		log.Fatal(err)
	}

	ws := listeners.NewWebsocket(listeners.Config{
		ID:        "ws1",
		Address:   *wsAddr,
		TLSConfig: wsTlsConfig,
	})
	err = server.AddListener(ws)
	if err != nil {
		log.Fatal(err)
	}

	stats := listeners.NewHTTPStats(
		listeners.Config{
			ID:      "info",
			Address: *infoAddr,
		},
		server.Info,
	)
	err = server.AddListener(stats)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		err := server.Serve()
		if err != nil {
			log.Fatal(err)
		}
	}()

	<-done
	server.Log.Warn("caught signal, stopping...")
	_ = server.Close()
	server.Log.Info("mochi mqtt shutdown complete")
}
