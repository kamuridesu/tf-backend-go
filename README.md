# TF Backend Go

A jump to the sky turns to a Rider Kick

## Intro

A simple HTTP backend to store Terraform/OpenTofu state information using SQLite or Postgres as database.

## Build

```
CGO_ENABLED=1 go build -ldflags='-s -w -extldflags "-static"' -o tf-backend
```
