# Fog Computing Assignment

## Prototyping Assignment Fog Computing Summer 2024

### Description

This project is part of a school assignment on fog computing. It includes a client, a router, and a watcher, all implemented in GoLang.

## Project Structure

- `client`: This directory contains the client implementation.
- `router`: This directory contains the router implementation.
- `watcher`: This directory contains the watcher implementation.
- `pkg/util/storage.go`: This file contains the `Storage` struct and its associated methods for managing clients.

## Storage

The `Storage` struct in `storage.go` is used to manage clients. It includes the following methods:

- `DeregisterInactiveClients(timeout time.Duration)`: This method deregisters any clients that have been inactive for longer than the specified timeout.

### Tasks

## Getting Started

To get started with this project, clone the repository and navigate to the project directory:

```bash
git clone https://github.com/yourusername/fog-computing-assignment.git
cd fog-computing-assignment
```
