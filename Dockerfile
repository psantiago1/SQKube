FROM golang:1.19

WORKDIR /go/src/github.com/SonarSource/helm-chart-sonarqube

COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download

RUN curl https://baltocdn.com/helm/signing.asc | gpg --dearmor | tee /usr/share/keyrings/helm.gpg > /dev/null &&\
    apt-get install apt-transport-https --yes &&\
    echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/helm.gpg] https://baltocdn.com/helm/stable/debian/ all main" | tee /etc/apt/sources.list.d/helm-stable-debian.list &&\
    apt-get update &&\
    apt-get install helm &&\
    helm repo add bitnami https://charts.bitnami.com/bitnami &&\
    helm repo add bitnami-pre2022 https://raw.githubusercontent.com/bitnami/charts/pre-2022/bitnami &&\
    helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx


