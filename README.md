# wakaexporter

Sync [wakatime](https://wakatime.com)'s [heartbeat](https://wakatime.com/developers#heartbeats) to postgresql,
and browse by grafana.

## Example of Grafana

[Dashboard](https://grafana.com/grafana/dashboards/14778/reviews).

![Dashboard Snaphot](https://grafana.com/api/dashboards/14778/images/10811/image)

## Usage

-   Install [Timescale](https://www.timescale.com/).
-   Create database by `schema/wakatime.sql`
-   Install wakaexporter by `go get github.com/j2gg0s/wakaexporter`
-   Sync hearbeats from wakatime periodic: `wakaexporter --api-key XXX --pg XXX`
