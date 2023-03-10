<div align="center">
  <br>
    <h1> 
      <img alt="Open Sauced" src="https://user-images.githubusercontent.com/31265908/215340095-9249933d-19c8-418c-b82c-ff911f577f9f.png" width="800px">
     </h1>
  <strong>Manage maintenance tasks performed during a working day</strong>
</div>
<br>
<p align="center">
  <a href="https://github.com/vitorbiten/maintenance-app/actions/workflows/lint-test-tag.yml">
    <img src="https://github.com/vitorbiten/maintenance-app/actions/workflows/lint-test-tag.yml/badge.svg" alt="Lint & Test & Tag" style="max-width: 100%;">
  </a>
  <a href="https://github.com/vitorbiten/maintenance-app/actions/workflows/build-deploy.yml">
    <img src="https://github.com/vitorbiten/maintenance-app/actions/workflows/build-deploy.yml/badge.svg" alt="Build & Deploy" style="max-width: 100%;">
  </a>
</p>
<p align="center">
  <img src="https://img.shields.io/badge/License-MIT-yellow.svg?style=flat" alt="License MIT">
  <a href="https://github.com/vitorbiten/maintenance-app/tags">
    <img src="https://img.shields.io/github/v/tag/vitorbiten/maintenance-app.svg?style=flat" alt="GitHub Tag">
  </a>
</p>

This app was created and deployed in a week to showcase [my](https://www.linkedin.com/in/vitorbiten/) back-end skills.
The premise is a maintenance app where technicians can create and update tasks, and managers can delete and see all the tasks, 
getting notified when new ones arrive. I hope you like it!


## Contents

- [Introduction](#introduction)
  - [About](#-about)
  - [Permissions Map](#-permissions-map)
- [Local development](#local-development)
  - [Prerequisites](#-prerequisites)
  - [Run](#%EF%B8%8F-run)
  - [Code Linting](#-code-linting)
  - [Tests and Coverage](#-tests-and-coverage)
  - [Performance Test](#-performance-test)
  - [Docs](#-docs)
  - [Kubernetes](#-kubernetes)
- [Deployment](#deployment)
  - [Setup](#-setup)
  - [CI](#-cd)
  - [CD](#-ci)
  
# Introduction

## ⭐ About

The task summary can contain up to 2500 characters of personal information, so it's encrypted with an aes 256 symmetric key.
It also contains the date when it was performed, the default will be now() if not informed.

When the task is created in the API, a message with the technician's nickname, task id, task date and manager email will be sent through rabbitmq to a Worker.
If there's more than one manager then mutiple messages will be sent.

Users need to be authenticated to perform actions, the api comes with basic user routes and the token id is checked thoroughly.
    
## 🔑 Permissions Map

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

It can be also be hosted on a [local kubernetes](#-kubernetes) cluster with the standart kubernetes dashboard.

## 📋 Prerequisites

In order to run the project from a container we need `make`, `docker`, `docker-compose` and `go>=1.18` installed.

## 🖥️ Run

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
go get -v github.com/rubenv/sql-migrate/...
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

If all goes well, you should see this:

```shell
maintenance_api       | 2023/01/29 23:14:27 --------------- Maintenance API ---------------
maintenance_api       | 2023/01/29 23:14:27 Listening to port :8080
maintenance_worker    | 2023/01/29 23:11:16 -------------- Maintenance Worker -------------
maintenance_worker    | 2023/01/29 23:11:16  [*] Waiting for messages.
```

This means the API has connected to the mysql db and the worker connected to the rabbitmq server.

You can manage the mysql database on the included phpmyadmin at http://localhost:9090/ and rabbitmq http://localhost:15672/ (guest:guest).

This is how the db is initially seeded:

<div align="center">
  <img src="https://user-images.githubusercontent.com/31265908/215361322-666a6c1c-0c0d-488c-9c49-4064389ddaaf.png" alt="PHPAdmin"/>
</div>

If you wanna keep the database contents, disable the seed line at the API's [main.go](/api/app/main.go.yml) file.

To stop the running containers:

```shell
make stop
```

## 🎨 Code linting

The project uses [golangci-lint](https://golangci-lint.run/) with standart linters config.
To check the code and styles quality, use the command:

```shell
npm run lint
```

## 🧪 Tests and coverage

To run the test suite, use the command:

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

## 📖 Performance Test

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

## 📖 Docs

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
export PATH=$(go env GOPATH)/bin:$PATH
```

After generation you can read the docs at http://localhost:8080/swagger/index.html, this is the overview:

<div align="center">
  <img src="https://user-images.githubusercontent.com/31265908/215364943-fc3a026a-64d0-4071-bdd9-cc236d3097d4.png" alt="Swagger docs"/>
</div>

Production docs are available [here ↗](http://139.144.159.246/swagger/index.html#)

## ⎈ Kubernetes

The project uses [minikube](https://github.com/kubernetes/minikube) as a local kubernetes cluster, you just need to run:

```shell
make deploy
```

It will start minikube, connect the Docker CLI to the docker daemon inside the minikube VM, then kubectl apply these external configs:

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
- Worker (worker.yaml)
  - Deployment 
- MySQL (mysql.yaml)
  - Deployment 
  - NodePort Service
  - Persistent Volume
  - Persistent Volume Claim
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

## 🔨 Setup

The first action "Deploy K8s" ([deploy-k8s.yml](/.github/workflows/deploy-k8s.yml)) uses Azure/k8s-set-context@v3 to set cluster context 
with KUBE_CONFIG before running kubectl commands deploying the base services.
For our production deploy we don't need to run kubernetes-dashboard because it comes pre-installed on the Linode cluster, we will kubectl apply this extra file:

- Linode Service (load-balancer.yaml)
  - LoadBalancer

This action needs to be triggered manually and be ran from a tag version so the correct images can be pulled from DockerHub.

## 🚥 CI

The second action "Lint, Test and Tag" ([lint-test-tag.yml](/.github/workflows/lint-test-tag.yml)) has three steps:

<div align="center">
  <img src="https://user-images.githubusercontent.com/31265908/215362147-f52280c1-beb7-4820-84b0-650ef9920c8c.png" alt="Lint, Test and Tag"/>
</div>

It runs the lint, tests suite, and then in case of a push/merge to main, [Github Tag Action](https://github.com/anothrNick/github-tag-action) bumps 
the version using #major, #minor (default), #patch or #none tags in the commit message. After the bump a tag is created, triggering the Build and Deploy workflow.

## ☁ CD	

The third action "Build and Deploy" ([build-deploy.yml](/.github/workflows/build-deploy.yml)) has two steps:

<div align="center">
  <img src="https://user-images.githubusercontent.com/31265908/215362806-b5dffc9d-1f10-47c3-8f8f-9cf3aec20f15.png" alt="Build and Deploy"/>
</div>

On a new tag created, it builds the image using this tag, pushes it to Dockerhub and applies the new deployment on the cluster.

### Check out the production api [here!](http://139.144.159.246/)

**[⬆ back to top](#contents)**
