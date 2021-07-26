![Empolis Grafana](docs/logo-horizontal.png)

[![License](https://img.shields.io/github/license/empolis/grafana)](LICENSE)

This repo is a fork of https://github.com/grafana/grafana as used by [Empolis Information Management GmbH](https://empolis.com) as part of [Empolis Service Express Industrial Analytics](https://www.service.express/industrial-analytics/).

The main changes compared to vanilla Grafana are:
* Support for generating PDFs of dashboards (via a forked grafana/grafana-image-renderer)
* Data source permissions (permissions fetched from an external API)
* Injection of dashboard timezone via variable
* User creation for unknown JWT users
* Extended OAuth2/JWT attribute handling
