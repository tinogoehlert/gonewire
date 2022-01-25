package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tinogoehlert/gonewire"
)

var (
	dir               = flag.String("path", "/sys/bus/w1/devices/", "-path /sys/bus/w1/devices/")
	addr              = flag.String("address", ":7777", "-address :7777")
	temperatureMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "temperature",
	}, []string{"sensor_id", "type"})
	readCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "sensor_read",
	}, []string{"sensor_id", "type"})
	errorCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "sensor_error",
	}, []string{"sensor_id", "type"})
)

func init() {
	flag.Parse()
	prometheus.MustRegister(
		temperatureMetric,
		readCounter,
		errorCounter,
	)
}

func main() {

	gw, err := gonewire.New(*dir)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			select {
			case <-sigs:
				cancel()
				return
			case v := <-gw.Values():
				ftemp, _ := strconv.ParseFloat(v.Value, 64)
				ftemp = ftemp / 1000
				temperatureMetric.WithLabelValues(v.ID, v.Type).Set(ftemp)
				readCounter.WithLabelValues(v.ID, v.Type).Inc()
			}
		}
	}()

	gw.OnReadError(func(e error, s *gonewire.Sensor) {
		log.Printf("[ERR] %s: %s", s.ID(), err.Error())
		errorCounter.WithLabelValues(s.ID(), s.TypeString()).Inc()
	})

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(*addr, nil)
	}()

	gw.Start(ctx, 10*time.Second)
}
