# Gateway API Layer

## Purpose

* Single entry point
* Schema definition
* Field-level RBAC enforcement
* Request → domain orchestration

## Frameworks

* Go Lang
    * gqlgen
    * chi
    * dataloader

## Rules

* No business logic
* No DB access
* Thin resolvers only

## Features

* GraphQL API

## Dependencies

* Auth
* Config
* Core