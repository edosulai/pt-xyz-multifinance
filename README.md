# PT XYZ Multifinance

Aplikasi layanan keuangan modern yang dibangun dengan arsitektur mikroservis menggunakan Go, gRPC, dan RESTful API.

## Arsitektur Aplikasi

```mermaid
graph TD
    A[Client Applications] --> B[Load Balancer/Reverse Proxy]
    B --> C[HTTP Server gRPC-Gateway :8080]
    C --> D[gRPC Server :50051]
    D --> E[Service Layer]
    E --> F[Use Cases]
    F --> G[Repositories]
    G --> H[PostgreSQL DB]
    
    subgraph Services
    E --> E1[UserService]
    E --> E2[LoanService]
    end
    
    subgraph UseCases
    F --> F1[UserUseCase]
    F --> F2[LoanUseCase]
    end
    
    subgraph Repositories
    G --> G1[UserRepository]
    G --> G2[LoanRepository]
    end
```

## Komponen Utama

### 1. API Layer

- **HTTP Server (gRPC-Gateway)**
  - Port: 8080
  - Menyediakan REST API endpoints
  - Swagger UI untuk dokumentasi API
  - CORS support untuk akses dari web client

- **gRPC Server**
  - Port: 50051
  - Menggunakan Protocol Buffers
  - Mendukung streaming dan komunikasi biner yang efisien
  - Authentication middleware dengan JWT

### 2. Layanan Core

- **User Service**
  - Manajemen pengguna
  - Autentikasi dan otorisasi
  - Rate limiting untuk keamanan
  - JWT token management

- **Loan Service**
  - Manajemen pinjaman
  - Perhitungan kredit
  - Validasi data pinjaman

### 3. Database

- **PostgreSQL**
  - Reliable dan ACID compliant
  - Mendukung transaksi kompleks
  - Connection pooling untuk performa optimal

## Teknologi yang Digunakan

### 1. Go (Golang)

- Performa tinggi dan konsumsi memori rendah
- Concurrent programming yang mudah dengan goroutines
- Static typing untuk keamanan kode
- Kompilasi cepat dan hasil binary yang ringan

### 2. gRPC & Protocol Buffers

- Komunikasi antar service yang efisien
- Strongly typed contract dengan proto files
- Support untuk streaming
- Cross-platform dan language-agnostic

### 3. PostgreSQL

- ACID compliance untuk data finansial
- Reliable untuk transaksi kompleks
- Mature dan well-tested
- Extensive indexing support

### 4. Docker

- Konsistensi environment
- Mudah dalam deployment
- Scalability dan orchestration
- Isolasi komponen

## Fitur Keamanan

1. **JWT Authentication**
   - Token-based authentication
   - Refresh token mechanism
   - Secure token storage

2. **Rate Limiting**
   - Proteksi dari brute force
   - Resource allocation control
   - DDoS protection

3. **Database Security**
   - Connection pooling
   - Prepared statements
   - Encrypted connections

## Cara Menjalankan Aplikasi

### Menggunakan Docker

1. **Prerequisites:**
   - Docker
   - Docker Compose

2. **Langkah-langkah:**

```bash
# Clone repository
git clone https://github.com/edosulai/pt-xyz-multifinance.git
cd pt-xyz-multifinance

# Build dan jalankan dengan Docker Compose
docker-compose up --build
```

### Manual Setup

1. **Prerequisites:**
   - Go 1.23+
   - PostgreSQL
   - Make

2. **Langkah-langkah:**

```bash
# Clone repository
git clone https://github.com/edosulai/pt-xyz-multifinance.git
cd pt-xyz-multifinance

# Install dependencies
go mod download

# Setup database
psql -U postgres -c "CREATE DATABASE xyz_multifinance"
migrate -path migrations -database "postgresql://postgres:root@localhost:5432/xyz_multifinance?sslmode=disable" up

# Build aplikasi
make build

# Jalankan aplikasi
./bin/server
```

## API Endpoints

### REST API (HTTP - Port 8080)

- Swagger UI: [http://localhost:8080/swagger/](http://localhost:8080/swagger/)
- API Base URL: [http://localhost:8080/v1/](http://localhost:8080/v1/)

### gRPC (Port 50051)

- Service definitions dalam `/proto` directory
- Mendukung gRPC clients

## Monitoring & Logging

- **Logging:**
  - JSON formatted logs
  - Level-based logging (debug, info, warn, error)
  - Output ke file dan stdout

- **Metrics:**
  - Connection pool metrics
  - Request latency
  - Error rates

## Development

### Testing

```bash
# Jalankan unit tests
make test

# Setup test database
./test/setup_test_db.ps1
```

### Protobuf Generation

```bash
# Generate proto files
make proto
```

## Struktur Project

```plaintext
.
├── cmd/                    # Main application entry
├── configs/               # Configuration files
├── internal/              # Internal packages
│   ├── handler/          # HTTP/gRPC handlers
│   ├── model/            # Domain models
│   ├── repo/             # Data access layer
│   └── usecase/          # Business logic
├── migrations/            # Database migrations
├── pkg/                   # Shared packages
├── proto/                # Protocol buffer definitions
└── test/                 # Test utilities and fixtures
```

## Contributing

1. Fork repository
2. Create feature branch
3. Commit changes
4. Create pull request

## License

Copyright (c) 2025 PT XYZ Multifinance
