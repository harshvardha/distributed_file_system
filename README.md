# Distributed File System (GFS-like)

A basic distributed file system inspired by Google File System (GFS) implemented in Go.

## Architecture

The system consists of three main components:

### 1. Master Server
- Manages file namespace and metadata
- Tracks chunk locations across chunk servers
- Monitors chunk server health via heartbeat
- Coordinates file operations

### 2. Chunk Servers
- Store file chunks (64MB each)
- Handle chunk read/write operations
- Report status to master via heartbeat
- Support chunk replication

### 3. Client
- Provides CLI for file operations
- Splits files into chunks for upload
- Reassembles chunks for download
- Communicates with master and chunk servers

## Features

- **Chunk-based Storage**: Files are split into 64MB chunks
- **Replication**: Each chunk is replicated 3 times for fault tolerance
- **Distributed Storage**: Chunks distributed across multiple chunk servers
- **gRPC Communication**: Efficient RPC between all components

## Prerequisites

- Go 1.21 or higher
- Protocol Buffers compiler (protoc)
- gRPC Go plugin

## Installation

1. Install dependencies:
```bash
go mod download
```

2. Generate protobuf code:
```bash
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/dfs.proto
```

## Usage

### 1. Start Master Server
```bash
go run cmd/master/main.go
```

### 2. Start Chunk Servers
Start multiple chunk servers on different ports:
```bash
go run cmd/chunkserver/main.go -port 9001 -storage ./storage1
go run cmd/chunkserver/main.go -port 9002 -storage ./storage2
go run cmd/chunkserver/main.go -port 9003 -storage ./storage3
```

### 3. Use Client

**Upload a file:**
```bash
go run cmd/client/main.go upload -file /path/to/file.txt -name myfile.txt
```

**List files:**
```bash
go run cmd/client/main.go list
```

**Download a file:**
```bash
go run cmd/client/main.go download -name myfile.txt -output /path/to/output.txt
```

## Configuration

- **Chunk Size**: 64MB (configurable in `common/utils.go`)
- **Replication Factor**: 3 (configurable in `common/utils.go`)
- **Master Address**: localhost:8000 (configurable)

## Future Enhancements

- Master replication for high availability
- Automatic re-replication on chunk server failure
- Snapshot support
- Optimized append operations
- Garbage collection for deleted files
- Chunk migration and load balancing
