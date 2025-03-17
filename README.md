# Distributed Ratelimiter with Raft and Serf

This project is an experimental implementation of a **distributed rate limiter** leveraging the **Raft consensus algorithm** and **Serf** for service discovery. The goal of this project is to explore the usage, limitations, and potential applications of Raft in a production-like environment.

Raft is a distributed consensus algorithm designed for fault tolerance, ensuring that a group of distributed nodes can agree on a consistent state, even when some nodes might fail. In this project, Raft is used to coordinate the state of a distributed rate limiter, ensuring consistency and reliability in rate-limiting across nodes in a fault-tolerant manner.

Serf is used for **service discovery** to allow nodes to communicate with each other in the cluster.

## Installation
To build and run the project, Docker is recommended. Docker provides an easy way to spin up the distributed rate limiter and its dependencies.


### Option 1: Build and Run using Docker Compose

1. Clone this repository:
    ```bash
    git clone {repo}
    cd ratelimiter
    ```
2. Build and run the application using docker-compose:
    ```
    docker-compose up --build
    ```
   
    This command will:
    * Build the Go application using the Dockerfile in the root directory.
    * Set up a multi-node Docker environment as specified in the docker-compose.yml file.
    * Start 3 nodes (node-1, node-2, node-3) and expose relevant ports.

3. The services will be available on the following ports:
    * Node 1: 20001 (HTTP), 50001 (Raft), 50011 (Discovery)
    * Node 2: 20002 (HTTP), 50002 (Raft), 50012 (Discovery)
    * Node 3: 20003 (HTTP), 50003 (Raft), 50013 (Discovery)

### Option 2: Build and Run without Docker (Local Development) ###

1. Clone this repository:
    ```bash
    git clone {repo}
    cd ratelimiter
    ```

2. Install dependencies:
    ```bash
    make build
    ```

3. Run the application for each node by specifying unique configurations in separate terminal windows:
    * For Node 1:
        ``` bash
        NODE_ID="node-1" SERVER_PORT="20001" RAFT_PORT="50001" DISCOVERY_PORT="50011" CLUSTERS="127.0.0.1:50011,127.0.0.1:50012,127.0.0.1:50013" ./ratelimiter
		```

    * For Node 2:
        ``` bash
        NODE_ID="node-2" SERVER_PORT="20002" RAFT_PORT="50002" DISCOVERY_PORT="50012" CLUSTERS="127.0.0.1:50011,127.0.0.1:50012,127.0.0.1:50013" ./ratelimiter
		```

    * For Node 3:
       ``` bash
       NODE_ID="node-3" SERVER_PORT="20003" RAFT_PORT="50003" DISCOVERY_PORT="50013" CLUSTERS="127.0.0.1:50011,127.0.0.1:50012,127.0.0.1:50013" ./ratelimiter
	   ```
## Docker Configuration

### Dockerfile
The provided Dockerfile defines a multi-stage build for the Go application:

1. Stage 1 (Builder):
    * Uses the golang image to build the Go binary.
    * The Go modules are used to manage dependencies.
    * The Go application is built for Linux and amd64 architecture.

2. Stage 2 (Final Image):
    * Uses the minimal alpine image to run the Go binary.
    * Copies the compiled binary from the build stage.
    * Exposes ports for the application and the Raft protocol.
    * Specifies the command to run the ratelimiter binary.

### Docker Compose
The docker-compose.yml defines the services for running multiple nodes of the distributed rate limiter.

Each node:
* Is built using the Dockerfile and has unique configurations for:
    * `NODE_ID`: Unique identifier for the node.
    * `SERVER_PORT`: Port for HTTP requests.
    * `RAFT_PORT`: Port for Raft communication.
    * `DISCOVERY_PORT`: Port for service discovery.
    * `CLUSTERS`: A comma-separated list of nodes for Raft consensus.

## Usage

Once the system is up and running, you can interact with the rate limiter using the HTTP API.


### Increment Rate for a Client

To increment the rate count for a specific client, you can use the `POST` request:

```bash
curl -s -X POST "http://localhost:20001/rate/increment?client_id=client-1"
```

#### Example Response (Success)
```bash
{"result": true}
```

#### Example Response (Failure)

If the rate limit has been exceeded:
```bash
{"result": false}
```

### Check the Rate Limit for a Client

To check if a client is within the rate limit, you can use the GET request:

```bash
curl -s -X GET "http://localhost:20001/rate/check?client_id=client-1"
```

#### Example Response

```bash
{"remaining_quota":5,"result":true}
```