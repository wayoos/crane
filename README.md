crane
=====

Crane to manage docker

Docker is an incredible tools, but when I start to create a continus delivery platform,
I start with Fig but it don't fit exactly my requierment. So I decide to start writing an application
to manage my docker.

## Getting Started

Start crane server. (you should have a docker installed and configured)

```shell
crane s
```

Create a Dockerfile like the example below, stating an nginx server

```
FROM ubuntu:14.04
RUN apt-get install -y nginx
CMD ["nginx"]
EXPOSE 80
```

Go in the folder where the Dockerfile is and execute

```shell
crane up
```


## Installation

```shell
curl -L https://github.com/wayoos/crane/releases/download/0.0.2/crane-`uname -s`-`uname -m` > /usr/local/bin/crane; chmod +x /usr/local/bin/crane
```
