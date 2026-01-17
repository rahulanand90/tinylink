# TinyLink URL Shortener Architecture

## System Overview

TinyLink is a high-performance URL shortener service built with Go, designed to handle 10K requests/second with a read-heavy traffic pattern. The architecture follows modern microservices principles with containerization, horizontal scaling, and separation of concerns.

## Architecture Components

### Load Balancing Layer
- **Nginx**: Acts as the entry point, distributing traffic across multiple application instances
- Ports: 80/443 (HTTP/HTTPS)
- Handles SSL termination and initial request routing

### API Gateway Layer
- Provides rate limiting and API key authentication
- Port: 8080
- Protects backend services from abuse and unauthorized access

### Core URL Service
- Written in Go for high performance and concurrency
- Stateless design allows horizontal scaling
- Multiple instances running behind the API Gateway
- Port: 8000
- Handles URL creation, retrieval, and redirection logic

### Data Storage
- **PostgreSQL**: Primary persistent storage with master-replica setup
- Port: 5432
- Stores URL mappings, user data, and analytics information
- Master for writes, replica for read scaling

### Caching Layer
- **Redis**: In-memory cache for frequently accessed URLs
- Port: 6379
- Also used for rate limiting counters
- Reduces database load for hot URLs

### Analytics Pipeline
- Asynchronous processing of click events
- **Kafka**: Message queue for click tracking
- Port: 9092
- **Analytics Service**: Processes events from the queue
- Port: 8001

## Request Flows

### URL Creation (POST)
1. Client sends POST request to create short URL
2. Request passes through Nginx to API Gateway
3. API Gateway validates API key and checks rate limits
4. Request forwarded to available TinyLink service instance
5. Service generates short code and stores mapping in PostgreSQL
6. New mapping is cached in Redis
7. Short URL returned to client

### URL Redirect (GET)
1. Client requests short URL
2. Request passes through Nginx to API Gateway
3. API Gateway forwards to available TinyLink service instance
4. Service checks Redis cache for URL mapping
5. If cache miss, service queries PostgreSQL
6. Click event sent to Kafka asynchronously
7. Client redirected to original URL

## Deployment

The entire system is containerized using Docker with separate networks for external and internal communication, enhancing security and isolation.

## Scaling Strategy

- Horizontal scaling of TinyLink service instances based on load
- Redis cluster for cache scaling
- PostgreSQL read replicas for database read scaling
- Kafka partitioning for analytics pipeline scaling
