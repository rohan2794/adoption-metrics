# adoption-metrics

install: docker-compose up -d

    Three components will be deployed ():

    1. grafana/grafana-oss:8.5.5

    2. rohan0227/metrics:v8

    3. prom/prometheus:latest

   Note: for OS=linux and ARCH=amd64, metrics image tag 		   will be v8 and for Darwin arm tag will be test

Grafana: http://localhost:3000 (Cred:admmin/admin)

Prometheus: http://localhost:9090/

uninstall:docker-compose down


Ref: https://techviewleo.com/run-prometheus-and-grafana-using-docker-compose/

Ref: https://percona.community/blog/2021/07/21/create-your-own-exporter-in-go/
