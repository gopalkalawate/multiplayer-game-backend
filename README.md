
# 🕹️ Multiplayer Game Backend (Go)

A high-performance, real-time multiplayer game backend built in Go. This project focuses on **distributed systems design**, exploring the intersection of WebSocket orchestration, concurrency-safe state management, and matchmaking logic.

---

## 🚀 Key Features

* **Matchmaking:** Redis-backed queuing system with asynchronous match creation.
* **Real-time Communication:** WebSocket-based full-duplex messaging using the **Hub-Client pattern**.
* **Session Isolation:** Room-based logic ensuring game state is isolated by unique Match IDs.
* **Synchronized Game Loop:** Tick-based simulation (20Hz) for deterministic state updates.
* **Concurrency Safety:** Use of goroutines and channels for non-blocking I/O and backpressure control.
* **Hybrid Storage:** Redis for low-latency ephemeral data; SQLite for reliable persistence.

---

## 🏗️ Architecture Overview

The system is designed with a clear separation of concerns to allow for future horizontal scaling.

### System Flow

1. **Ingress:** Clients join the pool via REST API.
2. **Orchestration:** Matchmaking workers monitor the Redis queue and pair players.
3. **Establishment:** Clients upgrade to WebSockets once a `matchID` is assigned.
4. **Simulation:** The `GameManager` spawns a dedicated tick-loop for the match.

### Project Structure

```text
.
├── cmd/server/          # Application entrypoint
├── internal/
│   ├── socket/          # WebSocket hub, clients, & game loop logic
│   ├── matchmaking/     # Redis queue & matching algorithms
│   ├── storage/         # SQLite persistence layer
│   └── utils/           # Shared helpers & constants
├── migrations/          # Database schema versions
└── go.mod               # Dependencies

```

---

## 🛠️ Technical Deep Dive

### WebSocket Hub Design

To prevent "Slow Consumer" issues, each client connection is handled by two independent goroutines:

* **Read Pump:** Listens for incoming messages and pushes them to the Hub.
* **Write Pump:** Listens for messages on a buffered channel and sends them to the client.

### Tick-Based Simulation

The server-side loop operates on a fixed interval:


Every tick, the server:

1. Processes the input queue.
2. Updates physics/logic state.
3. Broadcasts a state snapshot to all participants in the `matchID` room.

---

## 🚦 Getting Started

### Prerequisites

* **Go** (v1.21+)
* **Redis** (Running on `localhost:6379`)

### Installation & Run

1. **Clone the repository:**
```bash
git clone https://github.com/your-username/game-backend-go.git
cd game-backend-go

```


2. **Install dependencies:**
```bash
go mod download

```


3. **Start the server:**
```bash
go run cmd/server/main.go

```



The backend will be available at `http://localhost:8001`.

---

## 🗺️ Roadmap & Limitations

While this is a robust learning tool, it currently has specific constraints:

* [ ] **Auth:** Implementation of JWT-based authentication.
* [ ] **Scalability:** Horizontal scaling using Redis Pub/Sub for cross-server communication.
* [ ] **Authority:** Transitioning from client-authoritative to **server-authoritative** movement.
