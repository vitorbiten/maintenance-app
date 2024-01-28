# Multi-Tier Web Application with CI/CD Pipeline and Kubernetes Deployment


This project is a simple multi-tier web application designed to showcase DevOps skills, including Dockerization, CI/CD pipeline implementation, and Kubernetes deployment. The application includes a back-end service (API server), messaging service (RabbitMQ), and a database.


## Contents

- [API Introduction](#api-introduction)
  - [About](#about)
  - [Permissions Map](#permissions-map)
- [Local development](#local-development)
  - [Prerequisites](#prerequisites)
  - [Run](#run)
  - [Code Linting](#code-linting)
  - [Tests and Coverage](#tests-and-coverage)
  - [Performance Test](#performance-test)
  - [Docs](#docs)
  - [Kubernetes](#kubernetes)
- [Deployment](#deployment)
  - [Setup](#setup)
  - [CI](#ci)
  - [CD](#cd)
  
# API Introduction

## About

The backend API is a maintenance tasks management system that has two entities, users and tasks. 

The task summary can contain up to 2500 characters of personal information, so it's encrypted with an aes 256 symmetric key.
It also contains the date when it was performed, the default will be now() if not informed.

When the task is created in the API, a message with the technician's nickname, task id, task date and manager email will be sent through rabbitmq to a Worker.
If there's more than one manager then mutiple messages will be sent, the messages then can be used by the worker to perform any action like sending a notification.

Users need to be authenticated to perform actions, the api comes with basic user routes and the token id is checked thoroughly.

The API serves as good development base with authentication, messaging, gracefull shutdown, data validation, hot reloading and many tools and features for a development environment. 

## Permissions Map

<div align="center">
  <table cellpadding="5">
    <tbody align="center">
      <tr>
        <td colspan="2">
        </td>
        <td  width="250" >
            <h2> Manager </h2>
        </td>
        <td  width="250">
            <h2> Technician </h2>
        </td>
      </tr>
      <tr>
        <td width="160" rowspan="2">
            <h3> Users <br></h3>
        </td>
        <td>
            <b> self <br></b>
        </td>
        <td>
            <b> ⭕ create <br></b>
            <b> ✅ read <br></b>
            <b> ✅ update <br></b>
            <b> ✅ delete <br></b>
        </td>
        <td>
            <b> ✅ <br></b>
        </td>
      </tr>
      <tr>
        <td>
            <b> others <br></b>
        </td>
        <td>
            <b> ✅ read technicians <br></b>
            <b> ⭕ read managers <br></b>
            <b> ⭕ update <br></b>
            <b> ⭕ delete <br></b>
        </td>
        <td>
            <b> ⭕ <br></b>
        </td>
      </tr>
      <tr>
        <td width="160" rowspan="2">
            <h3> Tasks <br></h3>
        </td>
        <td>
            <b> self <br></b>
        </td>
        <td>
            <b> ⭕ <br></b>
        </td>
        <td>
            <b> ✅ create <br></b>
            <b> ✅ read <br></b>
            <b> ✅ update <br></b>
            <b> ⭕ delete <br></b>
        </td>
      </tr>
      <tr>
        <td>
            <b> others <br></b>
        </td>
        <td>
            <b> ✅ read <br></b>
            <b> ⭕ update <br></b>
            <b> ✅ delete <br></b>
        </td>
        <td>
            <b> ⭕ <br></b>
        </td>
      </tr>
    </tbody>
  </table>
</div>
  
# Local development

To get a top level view of the local development environment check out the [Makefile](/Makefile) or run:

```shell
make help
```

The project can be hosted locally through docker-compose using the following resources:

`maintenance-api`, `maintenance-worker`, `mysql`, `rabbitmq`, `phpmyadmin`

It uses [Air](https://github.com/cosmtrek/air) for hot-reloading, for more info check [docker-compose.yml](/docker-compose.yml).

It can be also be hosted on a [local kubernetes](#kubernetes) cluster with the standart kubernetes dashboard.

## Prerequisites

In order to run the project from a container we need `make`, `docker`, `docker-compose` and `go>=1.18` installed.

## Run

First run this command to create the .env files from the .env.examples:

```shell
make env
```

Then, run this command to start the development servers with docker-compose:

```shell
make run
```

Install Sql-Migrate

```shell
go install github.com/rubenv/sql-migrate/...@latest
```

Run the following command to export .env variables:

```shell
export $(cat .env | egrep -v "(^#.*|^$)" | xargs)
```

Then run the migration:

```shell
make migrate-up
```

If you get "sql-migrate: No such file or directory", you may need to add $GOPATH/bin to $PATH:

```shell
export PATH=$(go env GOPATH)/bin:$PATH
```

You can manage the mysql database on the included phpmyadmin at http://localhost:9090/ and rabbitmq http://localhost:15672/ (guest:guest).

To stop the running containers:

```shell
make stop
```

## Code linting

The project uses [golangci-lint](https://golangci-lint.run/) with standart linters config.
To check the code and styles quality, use the command:

```shell
make lint
```

## Tests and coverage

To run the test suite, use the command (you may need to close the app mysql):

```shell
make test
```

You can request a coverage report with:

```shell
make coverage
```

The cover report html page will be opened:

<div align="center">
  <img src="https://user-images.githubusercontent.com/31265908/215358478-04a741f1-865c-4d74-b721-7fdd6865d0f2.png" alt="Coverage report"/>
</div>

## Performance Test

You can run performance tests on the api using [k6](https://github.com/grafana/k6) with the command:

```shell
make perf
```

Here are the results for my local run:

<div align="center">
  <img src="https://user-images.githubusercontent.com/31265908/215403690-d93d2f7e-3b3a-40d1-9816-2cf63d028292.png" alt="Coverage report"/>
</div>

You can also watch the tests live on grafana at http://localhost:3000/d/k6/k6-load-testing-results:

<div align="center">
  <img src="https://user-images.githubusercontent.com/31265908/215407972-81690ad0-bdd8-4bf7-8432-81d8619db805.png" alt="Coverage report"/>
</div>

## Docs

The project uses [swag](https://github.com/swaggo/swag) to convert Go annotations to Swagger Documentation, install with:

```shell
go install github.com/swaggo/swag/cmd/swag@latest
```

Then run the following command to generate the docs:

```shell
make swagger
```

If you get "swag: command not found", you may need to add $GOPATH/bin to $PATH:

```shell
export PATH=$PATH:$(go env GOPATH)/bin
```

After generation you can read the docs at http://localhost:8080/swagger/index.html, this is the overview:

<div align="center">
  <img src="https://user-images.githubusercontent.com/31265908/215364943-fc3a026a-64d0-4071-bdd9-cc236d3097d4.png" alt="Swagger docs"/>
</div>

Production docs are available [here ↗](http://139.144.247.5/swagger/index.html#)

## Kubernetes

When combined with Kubernetes Horizontal Pod Autoscaler (HPA), Dockerized applications become highly scalable and self-healing:

- **Scalability:** Docker containers provide a consistent environment, and Kubernetes HPA allows automatic scaling based on resource usage or custom metrics. As the application experiences varying workloads, HPA dynamically adjusts the number of running pods to handle the load efficiently.

- **Self-Healing:** In the event of a pod failure or degradation, Kubernetes detects the issue and automatically restarts the affected pod. Docker's containerization ensures that the application can quickly recover by spinning up a new instance of the failed container..

The project uses [minikube](https://github.com/kubernetes/minikube) as a local kubernetes cluster, after installing it and kubectl, you need to run:

```shell
minikube start
```

```shell
eval $(minikube -p minikube docker-env)
```

The command minikube docker-env returns a set of Bash environment variable exports to configure your local environment to re-use the Docker daemon inside the Minikube instance.

```shell
make deploy
```

It will start minikube, setup the metrics-server for scalability control, then kubectl apply these external configs:

- [Kubernetes Dashboard](https://kubernetes.io/docs/tasks/access-application-cluster/web-ui-dashboard/)
- [Local Path Provisioner](https://github.com/rancher/local-path-provisioner)
- [RabbitMQ Cluster Operator](https://www.rabbitmq.com/kubernetes/operator/operator-overview.html)

Then build the images and kubectl apply these files in the kubernetes folder:

- Dashboard (kubernetes-dashboard.yaml)
  - Service Account
  - ClusterRoleBinding
- API (api.yaml)
  - Deployment 
  - NodePort Service
- API HPA (api-hpa.yaml)
  - HorizontalPodAutoscaler 
- Worker (worker.yaml)
  - Deployment 
- Worker HPA (worker-hpa.yaml)
  - HorizontalPodAutoscaler 
- MySQL (mysql.yaml)
  - Deployment 
  - NodePort Service
  - Persistent Volume
  - Persistent Volume Claim
- MySQL ConfigMap (mysql-configmap.yaml)
  - ConfigMap
- RabbitMQ (rabbitmq.yaml)
  - User Secrets
  - RabbitmqCluster
- Secrets (secrets.yaml)

You can access the dashboard with:

```shell
make dashboard
```

It will print the dashboard url (may not be acessible right away on new deploy), generate and print a token, and run kubectl proxy:

<div align="center">
  <img src="https://user-images.githubusercontent.com/31265908/215359712-2674b3d3-f01b-4c15-8544-cbd095422f63.png" alt="Make dashboard"/>
</div>

If you get the message "bind: address already in use" run before trying again:

```shell
make kill-dashboard
```

Now you can get your API and MySQL url's with:

```shell
make api-url
make mysql-url
```

You can also forward RabbitMQ management port to localhost with:

```shell
make rabbitmq
```

# Deployment

The deployment is divided into 3 Github Actions:

## Setup

The first action "Deploy K8s" ([deploy-k8s.yml](/.github/workflows/deploy-k8s.yml)) uses Azure/k8s-set-context@v3 to set cluster context 
with KUBE_CONFIG before running kubectl commands deploying the base services.
For our production deploy we don't need to run kubernetes-dashboard because it comes pre-installed on the Linode cluster, we will kubectl apply this extra file:

- Linode Service (load-balancer.yaml)
  - LoadBalancer

This action needs to be triggered manually and be ran from a tag version so the correct images can be pulled from DockerHub.

## CI

The second action "Lint and Test" ([lint-test.yml](/.github/workflows/lint-test.yml)) has two steps:

It runs the lint and test suite when a pull requested is created.

## CD	

The third action "Deploy Pipeline" ([lint-test-bump-build-deploy.yml](/.github/workflows/lint-test-bump-build-deploy.yml)) has five steps:

In case of a push/merge to main, the lint and tests are ran, [Github Tag Action](https://github.com/anothrNick/github-tag-action) bumps 
the version using #major, #minor (default), #patch or #none tags in the commit message. After the bump a tag is created, it Builds the image and Deploy to the K8s.

### Check out the production api [here!](http://139.144.247.5/)

**[⬆ back to top](#contents)**
