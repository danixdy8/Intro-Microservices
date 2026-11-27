# Arquitectura de Microservicios Multilenguaje con SQLite y Docker

Implementación de cuatro APIs independientes en **Python (FastAPI)**, **Java (Spring Boot)**, **Go (Gin)** y **JavaScript (Express)**. Cada microservicio gestiona su propia base de datos **SQLite** aislada y persistida mediante volúmenes independientes en **Docker Compose**.

## Arquitectura

| Servicio | Lenguaje y Framework | Puerto | Base de datos SQLite | Volumen Docker |
| :--- | :--- | :---: | :--- | :--- |
| `users-api` | Python · FastAPI | `8000` | `/data/users.db` | `users_data` |
| `auth-api` | Java · Spring Boot | `8001` | `/data/auth.db` | `auth_data` |
| `orders-api` | Go · Gin | `8002` | `/data/orders.db` | `orders_data` |
| `products-api` | JavaScript · Express | `8003` | `/data/products.db` | `products_data` |

## Estructura del repositorio

- `docker-compose.yml`: Orquestación de los 4 contenedores, puertos y 4 volúmenes independientes.
- `users-api/`: API en Python/FastAPI (`Dockerfile`, `requirements.txt`, `app/`).
- `auth-api/`: API en Java/Spring Boot (`Dockerfile`, `pom.xml`, `src/`).
- `orders-api/`: API en Go/Gin (`Dockerfile`, `go.mod`, `main.go`).
- `products-api/`: API en JavaScript/Express (`Dockerfile`, `package.json`, `src/`).

---

## Ejecución con Docker Compose

### 1. Construir y levantar los 4 microservicios

```bash
docker compose up --build -d
```

### 2. Comprobar estado de los contenedores

```bash
docker compose ps
```

### 3. Visualizar logs

```bash
# Todos los servicios
docker compose logs -f

# Un servicio específico
docker compose logs -f users-api
docker compose logs -f auth-api
docker compose logs -f orders-api
docker compose logs -f products-api
```

### 4. Detener los servicios (conservando los volúmenes SQLite)

```bash
docker compose down
```

### 5. Detener y eliminar también los volúmenes

```bash
docker compose down -v
```

---

## Pruebas de Endpoints

### Users API — Python / FastAPI (`http://localhost:8000`)

```bash
# Estado del servicio
curl http://localhost:8000/health

# Registrar usuario
curl -X POST http://localhost:8000/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Ana Torres","email":"ana@example.com","password":"secret123","role":"engineer"}'

# Consultar lista de usuarios
curl http://localhost:8000/users

# Consultar usuario por ID
curl http://localhost:8000/users/1

# Actualizar usuario
curl -X PUT http://localhost:8000/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"Ana Torres Actualizada","email":"ana@example.com","role":"lead"}'

# Eliminar usuario
curl -X DELETE http://localhost:8000/users/1
```

### Auth API — Java / Spring Boot (`http://localhost:8001`)

Credenciales precargadas automáticamente en `auth.db`:
- `admin@example.com` / `admin123`
- `user@example.com` / `user123`

```bash
# Verificar estado de la API
curl http://localhost:8001/auth/status

# Login con credenciales válidas (HTTP 200)
curl -X POST http://localhost:8001/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin@example.com","password":"admin123"}'

# Login con credenciales inválidas (HTTP 401)
curl -X POST http://localhost:8001/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin@example.com","password":"incorrecta"}'

# Registrar nueva credencial en auth.db
curl -X POST http://localhost:8001/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"nuevo@example.com","password":"clave123"}'

# Consultar sesiones / registros de autenticación en auth.db
curl http://localhost:8001/auth/sessions
```

### Orders API — Go / Gin (`http://localhost:8002`)

```bash
# Estado del servicio
curl http://localhost:8002/health

# Crear pedido
curl -X POST http://localhost:8002/orders \
  -H "Content-Type: application/json" \
  -d '{"customer_name":"Carlos Vega","product_summary":"Teclado Mecánico","quantity":2,"total":238.00}'

# Consultar todos los pedidos
curl http://localhost:8002/orders

# Consultar pedido por ID
curl http://localhost:8002/orders/1

# Actualizar estado del pedido (PENDING, PROCESSING, SHIPPED, DELIVERED, CANCELLED)
curl -X PATCH http://localhost:8002/orders/1/status \
  -H "Content-Type: application/json" \
  -d '{"status":"SHIPPED"}'

# Eliminar pedido
curl -X DELETE http://localhost:8002/orders/1
```

### Products API — JavaScript / Express (`http://localhost:8003`)

```bash
# Estado del servicio
curl http://localhost:8003/health

# Registrar producto
curl -X POST http://localhost:8003/products \
  -H "Content-Type: application/json" \
  -d '{"name":"Monitor 27 QHD","description":"Panel IPS 144Hz","price":329.99,"stock":15,"category":"Monitores"}'

# Consultar lista de productos
curl http://localhost:8003/products

# Consultar producto por ID
curl http://localhost:8003/products/1

# Actualizar producto
curl -X PUT http://localhost:8003/products/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"Monitor 27 QHD Pro","description":"Panel IPS 165Hz","price":349.99,"stock":12,"category":"Monitores"}'

# Eliminar producto
curl -X DELETE http://localhost:8003/products/1
```

---

## Verificación de Persistencia (Volúmenes SQLite)

1. Registra datos en las 4 APIs (`POST /users`, `POST /auth/login`, `POST /orders`, `POST /products`).
2. Detén y elimina los contenedores sin borrar los volúmenes:
   ```bash
   docker compose down
   ```
3. Recrea y vuelve a levantar los contenedores:
   ```bash
   docker compose up -d
   ```
4. Consulta nuevamente cada servicio (`GET /users`, `GET /auth/sessions`, `GET /orders`, `GET /products`). Todos los registros permanecen intactos en `users_data`, `auth_data`, `orders_data` y `products_data`.
