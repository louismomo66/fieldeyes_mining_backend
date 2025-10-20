# Mining Finance System Backend

A comprehensive backend API for managing mining finance operations including income tracking, expense management, inventory control, and financial analytics.

## Features

- **User Authentication & Authorization**
  - JWT-based authentication
  - Role-based access control (Admin/Standard users)
  - Password reset with OTP
  - User profile management

- **Income Management**
  - Track mineral sales and income
  - Support for multiple mineral types (Gold, Copper, Cobalt, Diamond, Other)
  - Payment status tracking
  - Customer information management

- **Expense Management**
  - Categorized expense tracking
  - Supplier management
  - Payment status tracking
  - Expense analytics and reporting

- **Inventory Management**
  - Track mineral and supply inventory
  - Low stock alerts
  - Quantity management
  - Value tracking

- **Financial Analytics**
  - Financial summaries
  - Monthly data analysis
  - Expense category breakdowns
  - Profit/loss calculations

## Technology Stack

- **Language**: Go 1.24.1
- **Framework**: Gorilla Mux
- **Database**: PostgreSQL with GORM
- **Authentication**: JWT tokens
- **Password Hashing**: bcrypt

## Prerequisites

- Go 1.24.1 or higher
- PostgreSQL 12 or higher
- Git

## Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd mineral/backend
   ```

2. **Install dependencies**
   ```bash
   go mod tidy
   ```

3. **Set up environment variables**
   ```bash
   cp env.example .env
   # Edit .env with your configuration
   ```

4. **Set up PostgreSQL database**
   ```sql
   CREATE DATABASE mining_data;
   CREATE USER mining_user WITH PASSWORD 'your_password';
   GRANT ALL PRIVILEGES ON DATABASE mining_data TO mining_user;
   ```

5. **Run the application**
   ```bash
   go run cmd/api/main.go
   ```

The server will start on `http://localhost:8080`

## API Endpoints

### Authentication
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/signup` - User registration
- `POST /api/v1/auth/forgot-password` - Request password reset
- `POST /api/v1/auth/reset-password` - Reset password with OTP

### User Profile
- `GET /api/v1/profile` - Get user profile
- `PUT /api/v1/profile` - Update user profile

### Income Management
- `GET /api/v1/income` - Get all income records
- `POST /api/v1/income` - Create income record
- `GET /api/v1/income/{id}` - Get specific income record
- `PUT /api/v1/income/{id}` - Update income record
- `DELETE /api/v1/income/{id}` - Delete income record
- `GET /api/v1/income/range?start_date=YYYY-MM-DD&end_date=YYYY-MM-DD` - Get income by date range

### Expense Management
- `GET /api/v1/expense` - Get all expense records
- `POST /api/v1/expense` - Create expense record
- `GET /api/v1/expense/{id}` - Get specific expense record
- `PUT /api/v1/expense/{id}` - Update expense record
- `DELETE /api/v1/expense/{id}` - Delete expense record
- `GET /api/v1/expense/range?start_date=YYYY-MM-DD&end_date=YYYY-MM-DD` - Get expenses by date range
- `GET /api/v1/expense/breakdown` - Get expense breakdown by category

### Inventory Management
- `GET /api/v1/inventory` - Get all inventory items
- `POST /api/v1/inventory` - Create inventory item
- `GET /api/v1/inventory/{id}` - Get specific inventory item
- `PUT /api/v1/inventory/{id}` - Update inventory item
- `DELETE /api/v1/inventory/{id}` - Delete inventory item
- `GET /api/v1/inventory/low-stock` - Get low stock items
- `PATCH /api/v1/inventory/{id}/quantity` - Update item quantity

### Analytics
- `GET /api/v1/analytics/summary` - Get financial summary
- `GET /api/v1/analytics/monthly?year=YYYY` - Get monthly data
- `GET /api/v1/analytics/expense-breakdown` - Get expense breakdown

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | Database host | localhost |
| `DB_PORT` | Database port | 5432 |
| `DB_USER` | Database user | postgres |
| `DB_PASSWORD` | Database password | postgres |
| `DB_NAME` | Database name | mining_data |
| `JWT_SECRET` | JWT signing secret | your-secret-key |
| `PORT` | Server port | 8080 |

## Database Schema

The application uses the following main entities:

- **Users**: User accounts with authentication
- **Income**: Income transactions from mineral sales
- **Expenses**: Expense transactions for operations
- **Inventory**: Inventory items (minerals and supplies)

## Development

### Running Tests
```bash
go test ./...
```

### Building for Production
```bash
go build -o bin/api cmd/api/main.go
```

### Docker Support
```bash
# Build Docker image
docker build -t mining-finance-api .

# Run with Docker Compose
docker-compose up
```

## Security Features

- Password hashing with bcrypt
- JWT token authentication
- CORS protection
- Input validation
- SQL injection prevention (GORM)
- Role-based access control

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License.
