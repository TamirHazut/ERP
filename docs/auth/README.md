# Auth / RBAC Module

## Purpose

* Authentication (users, tokens, sessions)
* Authorization (roles, permissions, policies)
* Tenant resolution

## Frameworks

* Go Lang
    * jwt-go
    * casbin

## Data

* Redis → sessions, tokens, rate limits
* MongoDB → users, roles, policies

## Features

* JWT
* Rate limiting
* Tenant resolution

## Dependencies

* Config
* DB

This module is pure infrastructure.