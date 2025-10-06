# Kanban API Documentation

## Overview
API untuk aplikasi Kanban Board dengan sistem autentikasi JWT dan manajemen project/task.

## Quick Start

### 1. Setup Database
Pastikan MySQL berjalan dan buat database:
```sql
CREATE DATABASE gin_api;
```

### 2. Environment Variables
Buat file `.env` di folder backend:
```env
DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=root
DB_PASSWORD=admin123
DB_NAME=gin_api
```

### 3. Run Application
```bash
cd backend
go run main.go
```

Server akan berjalan di `http://localhost:8080`

## API Documentation

### Swagger UI
Kunjungi `http://localhost:8080/swagger/index.html` untuk dokumentasi interaktif lengkap.

### Endpoints

#### Authentication
- `POST /register` - Register user baru
- `POST /login` - Login dan dapatkan JWT token

#### Projects (Perlu Authorization Header)
- `GET /projects` - Dapatkan semua project
- `POST /projects` - Buat project baru

#### Tasks (Perlu Authorization Header)
- `POST /tasks` - Buat task baru
- `PUT /tasks/{id}` - Update status task

### Authorization
Untuk endpoint yang memerlukan autentikasi, tambahkan header:
```
Authorization: Bearer <jwt-token>
```

## Testing

### Unit Tests
```bash
# Run all tests
go test ./controllers/... -v

# Run specific test file
go test ./controllers/authController_test.go -v
```

**Note:** Unit tests memerlukan database yang terkonfigurasi untuk berjalan dengan benar.

### Manual Testing dengan Swagger
1. Jalankan aplikasi
2. Buka `http://localhost:8080/swagger/index.html`
3. Register user baru melalui `/register`
4. Login melalui `/login` untuk mendapatkan JWT token
5. Klik "Authorize" di Swagger UI dan masukkan: `Bearer <token>`
6. Test endpoint lain seperti create project/task

### Example API Calls

#### 1. Register
```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"username": "testuser", "password": "password123"}'
```

#### 2. Login
```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username": "testuser", "password": "password123"}'
```

#### 3. Create Project (dengan JWT)
```bash
curl -X POST http://localhost:8080/projects \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your-jwt-token>" \
  -d '{"name": "My First Project"}'
```

#### 4. Create Task (dengan JWT)
```bash
curl -X POST http://localhost:8080/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your-jwt-token>" \
  -d '{"title": "Complete feature", "status": "todo", "project_id": 1}'
```

## Project Structure
```
backend/
├── controllers/          # HTTP handlers
│   ├── authController.go
│   ├── projectController.go
│   ├── taskController.go
│   └── *_test.go        # Unit tests
├── models/              # Data models
│   ├── user.go
│   ├── project.go
│   ├── task.go
│   └── swagger.go       # Swagger models
├── config/              # Database config
├── routes/              # Route definitions
├── middlewares/         # JWT middleware
├── docs/                # Generated Swagger docs
└── main.go             # Application entry point
```

## Features Implemented
✅ User Registration & Login  
✅ JWT Authentication  
✅ Project Management  
✅ Task Management  
✅ Swagger Documentation  
✅ Unit Tests  
✅ CORS Support  
✅ Database Migration  

## Next Steps
- Implement integration tests dengan test database
- Add validation untuk input data
- Implement soft delete untuk models
- Add pagination untuk list endpoints
- Add search/filter functionality