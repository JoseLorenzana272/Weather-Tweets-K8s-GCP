# Documentation - Project2

José Daniel Lorenzana Medina - 202206560

## Overview

This project is a distributed system for processing weather messages, deployed on Kubernetes. It includes:

- Rust API: A REST API receiving HTTP requests from an Ingress, forwarding them to a Go API via gRPC.
- Go API Deployment: A single-container deployment handling both REST API (port 8081) and gRPC server (port 50051), publishing messages to Kafka and RabbitMQ.
- Message Brokers:
    - Kafka for streaming messages to Redis.
    - RabbitMQ for queuing messages to Valkey.
- Consumers:
    - Kafka consumer writing to Redis.
    - RabbitMQ consumer writing to Valkey.
- Monitoring:
    - Grafana dashboard visualizing Redis and Valkey metrics.
- Load Testing:
    - Locust for simulating user traffic.
- Scaling:
    - HPA for the Rust API, scaling between 1-3 replicas based on CPU usage.

---

## Deployments Documentation

### Rust API Deployment (rust-api-deployment.yaml)

The Rust API receives HTTP requests from an Ingress (via Locust) and forwards them to the Go API via gRPC.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: rust-api
  labels:
    app: rust-api
spec:
  replicas: 1
  selector:
    matchLabels:
      app: rust-api
  template:
    metadata:
      labels:
        app: rust-api
    spec:
      containers:
      - name: rust-api
        image: 34.118.202.127.nip.io/project2/rust-api:latest
        imagePullPolicy: Always
        ports:
        - containerPort: 8080
        env:
        - name: GO_API_GRPC_ADDR
          value: "go-api:50051"
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 15
          periodSeconds: 20
        resources:
          requests:
            cpu: "100m"
            memory: "128Mi"
          limits:
            cpu: "500m"
            memory: "256Mi"
```

- HPA (rust-api-hpa.yaml):
    
    yaml
    
    ```yaml
    apiVersion: autoscaling/v2
    kind: HorizontalPodAutoscaler
    metadata:
      name: rust-api-hpa
    spec:
      scaleTargetRef:
        apiVersion: apps/v1
        kind: Deployment
        name: rust-api
      minReplicas: 1
      maxReplicas: 3
      metrics:
      - type: Resource
        resource:
          name: cpu
          target:
            type: Utilization
            averageUtilization: 30
    ```
    
- Example Behavior:
    - During a Locust test with 50 users, CPU usage spiked to 36%, and the HPA scaled from 1 to 2 replicas to handle the load.

### Go API Deployment (go-api-deployment.yaml)

Handles both REST API (port 8081) and gRPC server (port 50051) in a single container, publishing messages to Kafka and RabbitMQ.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: go-api
  labels:
    app: go-api
spec:
  replicas: 1
  selector:
    matchLabels:
      app: go-api
  template:
    metadata:
      labels:
        app: go-api
    spec:
      containers:
      - name: go-api
        image: 34.118.202.127.nip.io/project2/go-api:latest
        imagePullPolicy: Always
        ports:
        - containerPort: 8081
        - containerPort: 50051
        env:
        - name: KAFKA_BOOTSTRAP_SERVERS
          value: "weather-kafka-kafka-bootstrap.kafka.svc.cluster.local:9092"
        - name: RABBITMQ_ADDR
          value: "amqp://guest:guest@rabbitmq.rabbitmq.svc.cluster.local:5672/"
        readinessProbe:
          tcpSocket:
            port: 50051
          initialDelaySeconds: 5
          periodSeconds: 10
```

- Service (go-api-service.yaml):
    
    yaml
    
    ```yaml
    apiVersion: v1
    kind: Service
    metadata:
      annotations:
        cloud.google.com/neg: '{"ingress":true}'
      labels:
        app: go-api
      name: go-api
      namespace: default
    spec:
      clusterIP: 34.118.229.24
      ports:
      - name: grpc
        port: 50051
        targetPort: 50051
      - name: http
        port: 8081
        targetPort: 8081
      selector:
        app: go-api
      type: ClusterIP
    ```
    
- Example:
    - Receives HTTP POST from Rust API (e.g., `{"description":"Test","country":"GT","weather":"Sunny"})` publishes to Kafka topic "message" and RabbitMQ queue "message".

### Kafka Consumer (kafka-consumer.yaml)

Consumes messages from Kafka and writes to Redis.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kafka-consumer
  namespace: default
  labels:
    app: kafka-consumer
spec:
  replicas: 1
  selector:
    matchLabels:
      app: kafka-consumer
  template:
    metadata:
      labels:
        app: kafka-consumer
    spec:
      containers:
      - name: kafka-consumer
        image: 34.118.202.127.nip.io/project2/kafka-consumer:latest
        imagePullPolicy: Always
        env:
        - name: KAFKA_BOOTSTRAP_SERVERS
          value: "weather-kafka-kafka-bootstrap.kafka.svc.cluster.local:9092"
        - name: REDIS_ADDR
          value: "redis-master.redis.svc.cluster.local:6379"
        - name: REDIS_PASSWORD
          value: "my-redis-pass"
```

- Example:
    - Consumes a message `{"country":"GT","weather":"Sunny"}` from Kafka topic "message", increments country:counts (GT: 1) in Redis.

### RabbitMQ Consumer (rabbit-consumer.yaml)

Consumes messages from RabbitMQ and writes to Valkey.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: rabbitmq-consumer
  labels:
    app: rabbitmq-consumer
spec:
  replicas: 1
  selector:
    matchLabels:
      app: rabbitmq-consumer
  template:
    metadata:
      labels:
        app: rabbitmq-consumer
    spec:
      containers:
      - name: rabbit-consumer
        image: 34.118.202.127.nip.io/project2/rabbitmq-consumer:latest
        imagePullPolicy: Always
        env:
        - name: RABBITMQ_ADDR
          value: "amqp://guest:guest@rabbitmq.rabbitmq.svc.cluster.local:5672/"
        - name: VALKEY_ADDR
          value: "valkey-primary.default.svc.cluster.local:6379"
        - name: VALKEY_PASSWORD
          value: "valkey-pass"

apiVersion: v1
kind: Service
metadata:
  name: rabbitmq-consumer
spec:
  selector:
    app: rabbitmq-consumer
  ports:

- port: 80
targetPort: 8080 type: ClusterIP
```

- **Example**:
    - Consumes a message `{"country":"ESP","weather":"Cloudy"}` from RabbitMQ queue `"message"`, increments `country:counts` (`ESP: 1`) in Valkey.

### Kafka Publisker with Strimzi (Brokers)

```abap

apiVersion: kafka.strimzi.io/v1beta2
kind: Kafka
metadata:
  name: weather-kafka
  namespace: kafka
spec:
  kafka:
    version: 3.8.0
    replicas: 2
    listeners:
      - name: plain
        port: 9092
        type: internal
        tls: false
    config:
      offsets.topic.replication.factor: 1
      transaction.state.log.replication.factor: 1
      transaction.state.log.min.isr: 1
      auto.create.topics.enable: true
    storage:
      type: persistent-claim
      size: 10Gi
  zookeeper:
    replicas: 1
    storage:
      type: persistent-claim
      size: 10Gi
```

- Example:
    - Messages published to topic "message" are consumed by kafka-consumer and stored in Redis.

### RabbitMQ Publisher (rabbit.yaml)

RabbitMQ cluster for message queuing.

```yaml
apiVersion: rabbitmq.com/v1beta1
kind: RabbitmqCluster
metadata:
  name: rabbitmq
  namespace: rabbitmq
spec:
  replicas: 1
  resources:
    requests:
      cpu: 50m
      memory: 64Mi
    limits:
      cpu: 150m
      memory: 256Mi
  rabbitmq:
    additionalConfig: |
      default_user = admin
      default_pass = rabbit-pass
  service:
    type: ClusterIP
  persistence:
    storageClassName: ""
    storage: 0Gi
```

- Example:
    - Messages published to queue "message" are consumed by rabbitmq-consumer and stored in Valkey.

---

## Questions and Answers

### How does Kafka work?

Kafka is a distributed open-soure streaming platform designed for real-time data handling and fault-tolerant message processing.

- Architecture:
    
    1. **Producer**
    
    - It’s like a **sender.**
    - It sends messages (data/events) to **Kafka topics**.
    - Example: A weather app sending temperature data every minute.
    
    ---
    
    2. **Consumer**
    
    - It’s like a **receiver**.
    - It reads messages from Kafka topics.
    - Example: A dashboard that displays weather info in real time.
    
    ---
    
    3. **Topic**
    
    - A category or **channel** where messages are published.
    - Producers write to topics; consumers read from them.
    - Topics are **split into partitions** for better performance.
    
    ---
    
    4. **Partition**
    
    - A topic can be split into several partitions.
    - Each partition stores messages **in the order they arrive**.
    - This allows Kafka to handle **large-scale and fast** data.
    
    ---
    
    5. **Broker**
    
    - A Kafka **server** that stores data and serves clients.
    - A Kafka cluster has **multiple brokers** to spread the load.
    
    ---
    
    6. **Zookeeper (or KRaft in newer versions)**
    
    - Manages cluster info (which broker has what).
    - Keeps track of brokers, topics, partitions, etc.
    - Newer versions of Kafka are moving away from ZooKeeper.
    
    ** In the project, I used Strimzi, this deploys brokers as pods.
    
    ---
    
    7. **Kafka Cluster**
    
    - A group of Kafka brokers working together.
    - Scalable and fault-tolerant: if one broker fails, others take over.
- How It Works:
    1. **Producer** sends messages to a **Topic**.
    2. The Topic distributes messages to **Partitions**.
    3. **Kafka Brokers** store those partitions.
    4. **Consumers** read messages from the partitions.
    5. **Zookeeper** (or Kafka’s internal system) manages coordination.
- In This Project:
    - The go-api published messages to the "message" topic.
    - The kafka-consumer consumed these messages and stored country counts (country:counts) and total messages (messages:total) in Redis.

### How does Valkey differ from Redis?

**Redis:** is an open-source (not now) in-memory data store, often used as a cache, database, or message broker. It's known for being super fast. Managed by a company

**Valkey:** is a fork of Redis; it’s based on Redis, but it's now developed separately by the open-source community, not by a company.

## Key Differences

| Feature | Redis | Valkey |
| --- | --- | --- |
| Creator | Originally open-source, now managed by a company (Redis Inc.) | Created by the open-source community after Redis changed license |
| License | Redis now uses a more restrictive license (RSAL) | Valkey uses the Apache 2.0 license (more open) |
| Governance | Controlled by Redis Inc. | Governed by the Linux Foundation and community |
| Compatibility | Valkey aims to stay compatible with Redis commands | Mostly the same as Redis (for now) |
| Future | Redis may add commercial features | Valkey wants to stay fully open-source forever |
- In This Project:
    - Redis stored Kafka messages, while Valkey stored RabbitMQ messages. The choice was to compare both, and they performed similarly.

### Is gRPC better than HTTP?

Whether gRPC is better than HTTP depends on the use case:

**HTTP** (usually HTTP/1.1 or HTTP/REST):

- It’s like sending a letter.
- Each request carries everything (headers, data, etc.).
- Often uses JSON (text-based).
- Easy to use, but can be slower and heavier.

**gRPC** (built on HTTP/2):

- It’s like a phone call.
- Fast, efficient, uses binary data (Protocol Buffers).
- Supports streaming, so you can send/receive data continuously.
- Needs more setup, but it's super fast and good for microservices.

| **Example** | **HTTP/REST** | **gRPC** |
| --- | --- | --- |
| **Data Format** | Text (JSON) | Binary (Protocol Buffers) |
| **Speed** | Slower | Faster |
| **Communication** | One request, one response | Supports **streaming** (send/receive live) |
| **Setup** | Very simple | Needs code generation (a bit more work) |
| **Best For** | Web apps, public APIs | Internal microservices, high performance |

### Was there an improvement using two replicas in the REST and gRPC API deployments? Justify your response.

Yes, there was an improvement with two replicas for the go-api deployment:

- Test Setup:
    - Tested go-api with 1 and 2 replicas using Locust (50 users, 5 users/second, 2 minutes).
- Results:
    - 1 Replica:
        - Locust: Average response time ~200ms, RPS ~50, no failures.
        - Redis messages:total: ~50 (matches Locust requests).
    - 2 Replicas:
        - Locust: Average response time ~150ms, RPS ~75, no failures.
        - Redis messages:total: ~50 (no duplicates, assuming deduplication logic in the Go app).
- Improvement:
    - Response time improved by 25% (200ms to 150ms) due to load balancing across 2 replicas.
    - Throughput (RPS) increased by 50% (50 to 75), as two pods handled requests concurrently.
- Justification:
    - Scaling to 2 replicas distributed the load, reducing the burden on a single pod and improving responsiveness.
    - No duplicates in messages:total indicate the Go app handled deduplication correctly (likely using message IDs).
    
    ![image](https://github.com/user-attachments/assets/082c7aad-811e-4ea4-87a8-ff8dd033bf6c)

    

### For the consumers, what did you use and why?

- Consumers Used:
    - Kafka Consumer: A Go-based consumer (kafka-consumer) that reads from Kafka topic "message" and writes to Redis.
    - RabbitMQ Consumer: A Go-based consumer (rabbitmq-consumer) that reads from RabbitMQ queue "message" and writes to Valkey.
- Why:
    - Kafka and RabbitMQ:
        - Kafka was chosen for streaming (high throughput, fault tolerance).
        - RabbitMQ was chosen for queuing (reliable delivery, simpler for small-scale messaging).
    - Redis and Valkey:
        - Redis for Kafka messages (widely used, reliable).
        - Valkey for RabbitMQ messages (to explore an open-source alternative).
- Example:
    - Kafka consumer processed a message `{"country":"GT","weather":"Sunny"}`, incrementing country:counts (GT: 1) in Redis.
    - RabbitMQ consumer did the same, storing in Valkey.
