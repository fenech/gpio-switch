package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stianeikeland/go-rpio/v4"
)

type config struct {
	Pin  int `required:"true"`
	Port int `default:"7000"`
}

func main() {
	var c config
	envconfig.MustProcess("gpioswitch", &c)

	pinStatus := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gpio_switch",
			Help: "GPIO pin status (1 = on, 0 = off)",
		},
		[]string{
			"pin_number_bcm",
			"hostname",
		},
	)
	hostname, _ := os.Hostname()
	var state float64

	prometheus.MustRegister(pinStatus)

	err := rpio.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer rpio.Close()

	p := rpio.Pin(c.Pin)
	p.Output()

	http.HandleFunc("/on", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			p.High()
			w.WriteHeader(200)
		}
	})
	http.HandleFunc("/off", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			p.Low()
			w.WriteHeader(200)
		}
	})

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
		<head><title>Node Exporter</title></head>
		<body>
		<h1>Node Exporter</h1>
		<p><a href="/metrics">Metrics</a></p>
		<p><a href="/json">JSON Metrics</a></p>
		</body>
		</html>`))
	})
	http.HandleFunc("/json", func(w http.ResponseWriter, r *http.Request) {
		json, _ := json.Marshal(state)
		w.Write(json)
	})

	go func() {
		for {
			state = float64(p.Read())
			pinStatus.With(prometheus.Labels{
				"pin_number_bcm": strconv.Itoa(c.Pin),
				"hostname":       hostname,
			}).Set(state)
			time.Sleep(10 * time.Second)
		}
	}()

	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(c.Port), nil))
}
