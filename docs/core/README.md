# Core Module

## Purpose

* Inventory
* Vendors
* Clients
* Personnel
* Business rules
* Multi-tenancy enforcement

## Rules

* The only module allowed to mutate core business data
* Owns domain models and invariants
* Emits domain events

## Dependencies

* DB
* Config