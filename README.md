# Fog Computing Assignment

## Prototyping Assignment Fog Computing Summer Semester 2024

### Description

This project is spawned by a university assignment on fog computing. More precisely, it involves implementing a reliable message transfer network protocol in a use case of our choice.

## Project Structure

- `cmd/client`: This directory contains the client service implementation.
- `cmd/router`: This directory contains the router service implementation.
- `cmd/ram_watcher`: This directory contains the watcher service implementation for monitoring ram.
- `cmd/cpu_watcher`: This directory contains the watcher service implementation for monitoring cpu.
- `pkg/util/storage.go`: This file contains the `Storage` struct and its associated methods for managing clients.
- `pkg/util/udp.go`: This file contains the UDP communication methods for reliable message transfer.
- `pkg/util/messages.go`: This file contains the messages that are transmitted on the application layer.
- `pkg/util/util.go`: This file contains some util functions.
- `pkg/util/messages.go`: This file contains the request buffer struct.

## Getting Started

To get started with this project, clone the repository and navigate to the project directory:

```bash
git clone https://github.com/numyalai/fog-computing-assignment.git
cd fog-computing-assignment
```

To locally execute the project, simply run `make` inside the repository's root folder; this will build and automatically start the `router,` `client,` `cpu_watcher`, and `ram_watcher` services. If you want to cause some stress, try running the `go-stress` project, found [here](https://github.com/numyalai/go-stress).

## Endpoints

The following endpoints are used by the implemented services.

### router
- HTTP endpoint on port `6001`
- UDP socket listening on port `5001`

The HTTP endpoint accepts any HTTP request containing a body. The body will be forwarded to the registered client, with the most CPU and RAM available.

### client
- HTTP endpoint listening on port `5002`

The HTTP endpoint accepts POST requests having the following go format:
```
type WatcherMessage struct {
	Memory MemoryData
	Cpu    CpuData
}

type MemoryData struct {
	Free  uint64
	Total uint64
}

type CpuData struct {
	Free  uint64
	Total uint64
}
```
otherwise in JSON format:
```
{
  "memory": {
    "free": <unsigned_integer_64bit_value>,
    "total": <unsigned_integer_64bit_value>
  },
  "cpu": {
    "free": <unsigned_integer_64bit_value>,
    "total": <unsigned_integer_64bit_value>
  }
}
```
