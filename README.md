# CO2 Monitor Exporter

This is a [Prometheus](https://prometheus.io/) exporter for some similar USB CO2 monitors:

* [AIRCO2NTROL MINI](https://www.tfa-dostmann.de/produkt/co2-monitor-airco2ntrol-mini-31-5006/)
* [AIRCO2NTROL COACH](https://www.tfa-dostmann.de/produkt/co2-monitor-airco2ntrol-coach-31-5009/)
* [CO2Mini](https://www.co2meter.com/collections/indoor-air-quality/products/co2mini-co2-indoor-air-quality-monitor)

It uses the Linux HIDRAW API (`/dev/hidraw0` etc.) to access the CO2 Monitor.

## Metrics

* `co2monitor_co2_ppm`
* `co2monitor_temp_celsius`
* `co2monitor_humidity_rh`

## Installation

```bash
go get github.com/markuslindenberg/co2monitor_exporter
```

## Usage

```bash
~/go/bin/co2monitor_exporter --device /dev/hidraw0
```

## Package

This repository includes a [Go package for the USB CO2 monitor](https://godoc.org/github.com/markuslindenberg/co2monitor_exporter/co2monitor).