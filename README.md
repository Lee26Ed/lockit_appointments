# Lockit Appointments API

A RESTful API for managing appointments, built with Go and PostgreSQL. The system handles business appointments with support for users, staff, services, and review functionality.

## System Overview

- **User Management**: Create and manage user accounts with role-based access control
- **Business Management**: Register and manage service-providing businesses
- **Staff Management**: Manage business staff members and their schedules
- **Services**: Define and manage business services
- **Appointments**: Book, track, and manage service appointments with status tracking
- **Reviews**: Allow users to leave reviews for appointments
- **Rate Limiting**: Built-in rate limiting to prevent API abuse
- **CORS Support**: Configured CORS for cross-origin requests
- **Response Compression**: Automatic gzip compression for supported clients

## Prerequisites

- **Go**: 1.25.0 or later
- **PostgreSQL**: 12 or later
- **migrate**: Database migration tool
- **bash**: For running setup scripts

### Installing Dependencies

```bash
# On macOS (with Homebrew)
brew install go postgresql migrate

# On Ubuntu/Debian
sudo apt-get install golang-go postgresql postgresql-contrib
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

## Getting Started

### 1. Environment Setup

Create a `.envrc` file in the project root with your database credentials:

```bash
export DB_DSN="postgres://lockit_user:your_password@localhost:5432/lockit_appointments?sslmode=disable"
export DB_NAME="lockit_appointments"
export DB_USER="lockit_user"
export DB_PASSWORD="your_password"
export DB_HOST="localhost"
export DB_PORT="5432"
```

### 2. Database Setup

Run the setup script to create the database and user:

```bash
make db/setup
```

This will:

- Create a PostgreSQL database named `lockit_appointments`
- Create a dedicated user `lockit_user`
- Set up necessary permissions

### 3. Run Migrations

Apply all database migrations:

```bash
make db/migrations/up
```

This creates the following tables and structures:

- `roles` - User roles (admin, staff, customer)
- `users` - System users
- `businesses` - Service-providing businesses
- `business_staff` - Staff members of businesses
- `staff_schedule` - Staff working schedules
- `services` - Business services offered
- `appointments` - Appointment bookings with status tracking
- `reviews` - User reviews for appointments
- `appointment_status` - ENUM type for appointment statuses

### 4. Start the API Server

```bash
make run
```

The API will start on `http://localhost:3000`.

## API Configuration

The API runs with the following default settings:

- **Port**: `3000`
- **Environment**: `development`
- **Rate Limiting**:
    - Enabled by default
    - RPS (Requests Per Second): `2`
    - Burst: `5` requests
- **CORS**: Trusted origin set to `http://localhost:9000`

## Available Make Commands

### Database Operations

```bash
# Run migrations up
make db/migrations/up

# Rollback last migration
make db/migrations/down

# Rollback to specific version
make db/migrations/down version=N

# Check current migration version
make db/migrations/version

# Force migration to specific version (use with caution)
make db/migrations/force version=N

# Setup initial database
make db/setup

# Connect to database with psql
make db/psql
```

### Application

```bash
# Run the API server
make run

# Create a new migration
make db/migrations/new name=migration_name
```

### Middlewares

The API includes several middleware layers:

1. **CORS Middleware** - Handles cross-origin requests
2. **Rate Limiting** - Protects API from abuse with IP-based rate limiting
3. **Logging** - Emits structured JSON logs for all requests with:
    - HTTP method and path
    - Client IP address
    - Status code
    - Response size
    - Request duration
4. **Gzip Compression** - Automatically compresses responses when client sends `Accept-Encoding: gzip`

### Error Handling

The API provides standardized error responses in JSON format:

- `400 Bad Request` - Invalid request data
- `404 Not Found` - Resource not found
- `409 Conflict` - Edit conflict (concurrent modification)
- `422 Unprocessable Entity` - Validation errors
- `429 Too Many Requests` - Rate limit exceeded (includes `Retry-After` header)
- `500 Internal Server Error` - Server errors

## Development

### Project Structure

```
cmd/api/
├── main.go              # Application entry point
├── server.go            # Server configuration
├── routes.go            # Route definitions
├── handlers/            # HTTP request handlers
├── types/               # Type definitions
└── utils/               # Utility functions

cmd/internal/
├── data/                # Data models and database operations
└── validator/           # Input validation

migrations/              # Database migration files
scripts/                 # Setup and utility scripts
```
