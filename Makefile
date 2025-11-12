#Iniciar entorno de desarrollo (con hot reload)   | `docker compose -f docker-compose.dev.yml up`          |
# Iniciar entorno de producción (build optimizado) | `docker compose -f docker-compose.prod.yml up -d`      |
# Ejecutar solo build de producción (para CI/CD)   | `docker build --target prod -t kali-invoice-service .` |
# Ejecutar tests dentro del contenedor dev         | `docker exec -it kali-invoice-dev go test ./...`       |
