# Timer Microservice

## Table of Contents
1. [Introduction](#introduction)
2. [Features](#features)
3. [Prerequisites](#prerequisites)
4. [Installation](#installation)
5. [Configuration](#configuration)
6. [Running the Application](#running-the-application)
7. [API Documentation](#api-documentation)
8. [WebSocket Protocol](#websocket-protocol)
9. [Deployment](#deployment)
10. [Testing](#testing)
11. [Monitoring and Logging](#monitoring-and-logging)
12. [Contributing](#contributing)
13. [License](#license)

## Introduction

The Timer Microservice is a Go-based application designed to manage multiple timers in parallel. It provides RESTful API endpoints and WebSocket connections for creating, pausing, resuming, stopping, and modifying timers. The service is designed with scalability and reliability in mind, featuring persistent storage with MySQL and Redis for fast data retrieval and timer state preservation across server restarts.

## Features

- Create, pause, resume, stop, and modify timers
- Real-time updates via WebSocket connections
- Persistent storage using MySQL
- Redis caching for improved performance
- Timer state preservation across server restarts
- Dockerized deployment
- Structured logging
- Graceful shutdown

## Prerequisites

- Go 1.17 or later
- MySQL 8.0 or later
- Redis 6.0 or later
- Docker and Docker Compose (optional, for containerized deployment)

## Installation

1. Clone the repository:
   ```
   git clone https://github.com/your-username/timer-microservice.git
   cd timer-microservice
   ```

2. Install dependencies:
   ```
   go mod tidy
   ```

## Configuration

The application uses environment variables for configuration. Create a `.env` file in the root directory with the following contents:

```
DATABASE_DSN=user:password@tcp(localhost:3306)/timer_db?charset=utf8mb4&parseTime=True&loc=Local
PORT=8080
REDIS_ADDR=localhost:6379
```

Adjust the values according to your environment.

## Running the Application

### Local Development

1. Ensure MySQL and Redis are running locally.
2. Set up the environment variables in the `.env` file.
3. Run the application:
   ```
   go run main.go
   ```

### Using Docker Compose

1. Make sure Docker and Docker Compose are installed on your system.
2. Run the following command in the project root:
   ```
   docker-compose up --build
   ```

This will start the application, MySQL, and Redis in separate containers.

## API Documentation

### Create Timer

- **URL**: `/timer`
- **Method**: `POST`
- **Request Body**:
  ```json
  {
    "sessionId": "string",
    "maxTime": number
  }
  ```
- **Response**:
  ```json
  {
    "id": number,
    "sessionId": "string",
    "currentTime": number,
    "maxTime": number,
    "isPaused": boolean
  }
  ```

### Pause Timer

- **URL**: `/timer/{id}/pause`
- **Method**: `PUT`
- **Response**: Updated timer object

### Resume Timer

- **URL**: `/timer/{id}/resume`
- **Method**: `PUT`
- **Response**: Updated timer object

### Stop Timer

- **URL**: `/timer/{id}/stop`
- **Method**: `PUT`
- **Response**:
  ```json
  {
    "message": "Timer stopped and deleted"
  }
  ```

### Modify Timer

- **URL**: `/timer/{id}/modify`
- **Method**: `PUT`
- **Request Body**:
  ```json
  {
    "maxTime": number
  }
  ```
- **Response**: Updated timer object

## WebSocket Protocol

### Customer WebSocket

- **URL**: `/ws/customer/{sessionID}`
- **Description**: Provides real-time updates for a specific customer's timer.

### Game Master WebSocket

- **URL**: `/ws/gamemaster/{sessionID}`
- **Description**: Provides real-time updates for all active timers.

### WebSocket Message Structure

```json
{
  "type": "MESSAGE_TYPE",
  "payload": {}
}
```

Message types:
- `TIMER_UPDATE`
- `TIMER_CREATE`
- `TIMER_PAUSE`
- `TIMER_RESUME`
- `TIMER_STOP`
- `TIMER_MODIFY`

## Deployment

### Using Docker

1. Build the Docker image:
   ```
   docker build -t timer-microservice .
   ```

2. Run the container:
   ```
   docker run -p 8080:8080 --env-file .env timer-microservice
   ```

### Using Kubernetes

1. Create a Kubernetes deployment and service:
   ```yaml
   apiVersion: apps/v1
   kind: Deployment
   metadata:
     name: timer-microservice
   spec:
     replicas: 3
     selector:
       matchLabels:
         app: timer-microservice
     template:
       metadata:
         labels:
           app: timer-microservice
       spec:
         containers:
         - name: timer-microservice
           image: your-registry/timer-microservice:latest
           ports:
           - containerPort: 8080
           envFrom:
           - configMapRef:
               name: timer-microservice-config
   ---
   apiVersion: v1
   kind: Service
   metadata:
     name: timer-microservice
   spec:
     selector:
       app: timer-microservice
     ports:
     - port: 80
       targetPort: 8080
   ```

2. Apply the configuration:
   ```
   kubectl apply -f kubernetes-config.yaml
   ```

## Testing

To run the tests:

```
go test ./...
```

## Monitoring and Logging

- The application uses structured logging with Zap logger.
- For production deployments, consider setting up:
  - Prometheus for metrics collection
  - Grafana for visualization
  - ELK stack or similar for log aggregation and analysis

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.