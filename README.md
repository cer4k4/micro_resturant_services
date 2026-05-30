# Micro Restaurant Services

A distributed microservices application for managing restaurant operations with order management, delivery tracking, and restaurant inventory. Built with **Go**, containerized with **Docker**, and monitored with **Zipkin** distributed tracing.

## 📋 Overview

This project demonstrates a modern microservices architecture implementing:
- **Hexagonal Architecture (Ports & Adapters)** for clean separation of concerns
- **Distributed Tracing** with Zipkin for observability
- **MongoDB** for data persistence
- **Docker** containerization for easy deployment
- **Inter-service Communication** with REST APIs

### Architecture

The system consists of three independent microservices:

```
┌─────────────────────────────────────────────────────────────┐
│                   Restaurant Services                        │
├──────────────────────┬──────────────────────┬────────────────┤
│  Restaurant Service  │   Order Service      │ Delivery Service│
│  (Port 8080)         │   (Port 8082)        │   (Port 8083)  │
│                      │                      │                │
│  • Menu Management   │  • Order Creation    │  • Track Orders│
│  • Availability      │  • Order Listing     │  • Assign Driver│
│  • Restaurant Info   │  • Order Status      │  • Update Status│
└──────────────────────┴──────────────────────┴────────────────┘
            ↓                    ↓                     ↓
        MongoDB              MongoDB                MongoDB
     (restaurantdb)        (orderdb)           (deliverydb)
```

## 🏗️ Architecture Pattern: Hexagonal (Ports & Adapters)

Each service follows **Hexagonal Architecture** principles:

```
┌─────────────────────────────────────────────────────┐
│              APPLICATION/DOMAIN LAYER               │
│         (Business Logic - Framework Agnostic)      │
├─────────────────────────────────────────────────────┤
│                                                     │
│  INBOUND PORTS          OUTBOUND PORTS             │
│  (Interfaces)           (Interfaces)               │
│    ↓                        ↓                       │
├─────────────────────────────────────────────────────┤
│                                                     │
│  INBOUND ADAPTERS       OUTBOUND ADAPTERS          │
│  (HTTP Handlers)        (Database, Services)       │
│    ↓                        ↓                       │
├─────────────────────────────────────────────────────┤
│                                                     │
│  EXTERNAL SYSTEMS (HTTP APIs, Databases, etc.)    │
│                                                     │
└─────────────────────────────────────────────────────┘
```

### Benefits of Hexagonal Architecture

✅ **Testability** - Business logic separated from frameworks  
✅ **Maintainability** - Clear dependency flow (inward)  
✅ **Flexibility** - Easy to swap implementations (DB, API, etc.)  
✅ **Framework Independence** - Core logic doesn't depend on Echo or Mongo  
✅ **Scalability** - Services are loosely coupled  

## 📁 Directory Structure

```
micro_resturant_services/
│
├── resturantservice/          # Restaurant Service
│   ├── main.go
│   ├── go.mod
│   ├── .env
│   ├── Dockerfile
│   │
│   ├── config/                # Configuration (Adapter)
│   │   └── config.go
│   │
│   ├── domain/                # Domain Layer
│   │   ├── model/
│   │   │   └── restaurant.go
│   │   └── service/           # Business Logic (Use Cases)
│   │       └── restaurant_service.go
│   │
│   ├── port/                  # Port Definitions
│   │   ├── input/             # Inbound Ports (Interfaces)
│   │   │   └── restaurant_handler.go
│   │   └── output/            # Outbound Ports (Interfaces)
│   │       └── restaurant_repository.go
│   │
│   └── adapter/               # Adapter Implementations
│       ├── handler/           # HTTP Handlers (Inbound)
│       │   └── restaurant_handler.go
│       └── repository/        # Database (Outbound)
│           └── mongo_repository.go
│
├── orderservice/              # Order Service (Same Structure)
│   └── ...
│
├── deliveryservice/           # Delivery Service (Same Structure)
│   └── ...
│
└── docker-compose.yml         # Orchestration
```

## 🚀 Quick Start

### Prerequisites

- Docker & Docker Compose
- Go 1.19+ (for local development)
- MongoDB
- Zipkin

### Installation & Running

#### Using Docker Compose (Recommended)

```bash
# Clone the repository
git clone https://github.com/cer4k4/micro_resturant_services.git
cd micro_resturant_services

# Start all services with monitoring
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

#### Services will be available at:
- **Restaurant Service**: http://localhost:8080
- **Order Service**: http://localhost:8082
- **Delivery Service**: http://localhost:8083
- **Zipkin UI**: http://localhost:9411

### Local Development

```bash
# Install dependencies for each service
cd resturantservice && go mod download
cd ../orderservice && go mod download
cd ../deliveryservice && go mod download

# Run individual services
cd resturantservice && go run main.go
cd orderservice && go run main.go
cd deliveryservice && go run main.go
```

## 📚 API Documentation

### Restaurant Service (Port 8080)

```
POST   /restaurants          # Create restaurant
GET    /restaurants          # List all restaurants
GET    /restaurants/:id      # Get restaurant details
PUT    /restaurants/:id      # Update restaurant
DELETE /restaurants/:id      # Delete restaurant
```

### Order Service (Port 8082)

```
POST   /orders               # Create order
GET    /orders               # List orders
GET    /orders/:id           # Get order details
PUT    /orders/:id           # Update order status
DELETE /orders/:id           # Cancel order
```

### Delivery Service (Port 8083)

```
POST   /deliveries           # Create delivery
GET    /deliveries           # List deliveries
GET    /deliveries/:id       # Get delivery details
PUT    /deliveries/:id       # Update delivery status
```

## 🔍 Distributed Tracing with Zipkin

Each service is instrumented with **OpenZipkin** for distributed tracing:

- **Access Zipkin UI**: http://localhost:9411
- **View service dependencies**: http://localhost:9411/zipkin/
- **Trace requests across services**: Search by trace ID

### Example Trace Flow

```
CreateOrder (Restaurant Service)
    ↓ [trace-id: abc123]
GetRestaurants (Order Service)
    ↓ [trace-id: abc123, span-id: xyz789]
CreateDelivery (Delivery Service)
    ↓ [trace-id: abc123, span-id: def456]
SaveToDatabase (MongoDB)
```

## 🏛️ Hexagonal Architecture Details

### Layer Responsibilities

#### 1. **Domain Layer** (`domain/`)
- **Models**: Pure data structures (Restaurant, Order, Delivery)
- **Services**: Business logic, use cases, domain rules
- **No Dependencies**: On frameworks, external libraries, or adapters

```go
// Example: Domain Service (Framework Independent)
func (s *RestaurantService) CreateRestaurant(ctx context.Context, restaurant *Restaurant) error {
    if restaurant.Name == "" {
        return errors.New("restaurant name is required")
    }
    return s.repository.Save(ctx, restaurant)
}
```

#### 2. **Port Layer** (`port/`)
- **Input Ports**: Handler interfaces that external adapters must implement
- **Output Ports**: Repository/Service interfaces that domain needs

```go
// Input Port (Inbound)
type RestaurantHandler interface {
    CreateRestaurant(ctx context.Context, req CreateRestaurantRequest) error
    ListRestaurants(ctx context.Context) ([]Restaurant, error)
}

// Output Port (Outbound)
type RestaurantRepository interface {
    Save(ctx context.Context, restaurant *Restaurant) error
    FindAll(ctx context.Context) ([]Restaurant, error)
}
```

#### 3. **Adapter Layer** (`adapter/`)
- **Inbound Adapters**: HTTP handlers (Echo framework)
- **Outbound Adapters**: Database drivers (MongoDB), external service clients

```go
// HTTP Handler Adapter (Framework Specific)
func (h *RestaurantHandler) CreateRestaurant(c echo.Context) error {
    var req CreateRestaurantRequest
    if err := c.Bind(&req); err != nil {
        return c.JSON(400, map[string]string{"error": "invalid request"})
    }
    return h.service.CreateRestaurant(c.Request().Context(), &req.ToModel())
}

// MongoDB Repository Adapter
func (r *MongoRepository) Save(ctx context.Context, restaurant *Restaurant) error {
    collection := r.db.Database("restaurantdb").Collection("restaurants")
    _, err := collection.InsertOne(ctx, restaurant)
    return err
}
```

## 🧪 Testing Strategy

Hexagonal Architecture enables **easy unit testing**:

```go
// Mock Repository for Testing
type MockRestaurantRepository struct{}

func (m *MockRestaurantRepository) Save(ctx context.Context, r *Restaurant) error {
    // Mock implementation
    return nil
}

// Test Business Logic Without Framework
func TestCreateRestaurant(t *testing.T) {
    service := RestaurantService{repository: &MockRestaurantRepository{}}
    err := service.CreateRestaurant(context.Background(), &Restaurant{Name: "Test"})
    assert.NoError(t, err)
}
```

## 🔧 Configuration

Each service requires a `.env` file:

```env
# Restaurant Service
SERVER_PORT=8080
MONGO_URI=mongodb://mongo:27017
DATABASE_NAME=restaurantdb
ZIPKIN_ENDPOINT=http://zipkin:9411/api/v2/spans

# Order Service
SERVER_PORT=8082
MONGO_URI=mongodb://mongo:27017
DATABASE_NAME=orderdb
ZIPKIN_ENDPOINT=http://zipkin:9411/api/v2/spans

# Delivery Service
SERVER_PORT=8083
MONGO_URI=mongodb://mongo:27017
DATABASE_NAME=deliverydb
ZIPKIN_ENDPOINT=http://zipkin:9411/api/v2/spans
```

## 📊 Service Communication

Services communicate synchronously via REST APIs:

```
Order Service (CreateOrder)
    ↓ HTTP GET
Restaurant Service (GetRestaurant)
    ↓ Response with restaurant details
Order Service (SaveOrder)
    ↓ HTTP POST
Delivery Service (CreateDelivery)
    ↓ Save to database
```

## 🔐 Best Practices Implemented

✅ **Separation of Concerns** - Domain, Ports, Adapters  
✅ **Dependency Inversion** - Depend on interfaces, not implementations  
✅ **Single Responsibility** - Each layer has one reason to change  
✅ **Open/Closed Principle** - Open for extension, closed for modification  
✅ **Interface Segregation** - Small, focused interfaces  
✅ **Observability** - Distributed tracing with Zipkin  
✅ **Containerization** - Docker for consistent environments  

## 🚨 Error Handling

Each service implements proper error handling:

```go
// Domain errors
type ErrorType string

const (
    ErrRestaurantNotFound ErrorType = "RESTAURANT_NOT_FOUND"
    ErrInvalidInput       ErrorType = "INVALID_INPUT"
    ErrDatabaseError      ErrorType = "DATABASE_ERROR"
)

type DomainError struct {
    Type    ErrorType
    Message string
}
```

## 📈 Monitoring & Logging

- **Zipkin**: Distributed tracing across services
- **Echo Middleware**: Request logging and recovery
- **Structured Logging**: JSON logs for better analysis

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## 📄 License

This project is open source and available under the MIT License.

## 📞 Support

For issues and questions:
- GitHub Issues: [Project Issues](https://github.com/cer4k4/micro_resturant_services/issues)
- Email: [Your Email]

---

**Built with ❤️ using Hexagonal Architecture principles**
