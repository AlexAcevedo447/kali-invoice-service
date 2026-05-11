GUIA RAPIDA - KALI AUTH CONTEXT (kali-auth-context)

1) DESCOMPRIMIR Y ENTRAR AL PROYECTO
unzip kali-auth-context.zip
cd kali-auth-context

2) CREAR ARCHIVO DE ENTORNO LOCAL
cp .env.example .env

3) CONFIGURAR VARIABLES MINIMAS EN .env
DB_HOST=localhost
DB_PORT=5432
DB_USER=tu_usuario
DB_PASSWORD=tu_password
DB_NAME=kali_auth_dev
JWT_SECRET=un_secreto_largo_y_seguro

Opcional (crear usuario master al arrancar):
SEED_MASTER_ENABLED=true
SEED_MASTER_PASSWORD=TuPasswordSegura123!

4) LEVANTAR POSTGRESQL (OPCION RAPIDA CON DOCKER)
docker run -d --name kali-auth-db \
	-e POSTGRES_USER=tu_usuario \
	-e POSTGRES_PASSWORD=tu_password \
	-e POSTGRES_DB=postgres \
	-p 5432:5432 \
	postgres:16

Nota:
- La aplicacion intenta crear la base kali_auth_dev automaticamente si no existe.
- El usuario debe poder conectarse a la base postgres y tener permiso CREATEDB.

5) DESCARGAR DEPENDENCIAS
go mod download

6) EJECUTAR LA API
go run ./cmd/api

7) TROUBLESHOOTING RAPIDO
- Si el puerto 5432 esta ocupado, cambia DB_PORT en .env y publica ese puerto en Docker.
- Si falla autenticacion de BD, revisa DB_USER y DB_PASSWORD.
- Si aparece error al generar token, revisa que JWT_SECRET no este vacio.
- Si no se crea el master user, valida que SEED_MASTER_ENABLED=true y que tenga password.

8) URL LOCAL
http://localhost:8080


━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

GUIA RAPIDA - KALI INVOICE SERVICE (kali-invoice-service)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 SERVICIO: Gestión de facturas (Go + Fiber + PostgreSQL)
 PUERTO:   8081 (por defecto en .env.example)
 BD:       kali_invoices (se crea automáticamente si no existe)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

1) DESCOMPRIMIR Y ENTRAR AL PROYECTO
   unzip kali-invoice-service.zip
   cd kali-invoice-service

2) CREAR ARCHIVO DE ENTORNO LOCAL
   cp .env.example .env

3) CONFIGURAR VARIABLES EN .env
   APP_PORT=8081            # cambia si el puerto está ocupado
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=tu_usuario
   DB_PASSWORD=tu_password
   DB_NAME=kali_invoices_dev
   DB_SSLMODE=disable
   RABBITMQ_ENABLED=false   # no es necesario para desarrollo

4) LEVANTAR POSTGRESQL (OPCION RAPIDA CON DOCKER)
   docker run -d --name kali-invoice-db \
     -e POSTGRES_USER=tu_usuario \
     -e POSTGRES_PASSWORD=tu_password \
     -e POSTGRES_DB=postgres \
     -p 5432:5432 \
     postgres:16

   Nota:
   - La app crea la BD kali_invoices_dev automáticamente si no existe.
   - El usuario debe poder conectarse a la BD "postgres" y tener permiso CREATEDB.

5) DESCARGAR DEPENDENCIAS
   go mod vendor

6) EJECUTAR LA API
   go run ./cmd/api

7) TROUBLESHOOTING RAPIDO
   - Puerto ocupado: cambia APP_PORT en .env (ej: 8082, 8083...).
   - Puerto 5432 ocupado: cambia DB_PORT en .env y el -p del docker run.
   - Error de autenticación BD: revisa DB_USER y DB_PASSWORD.
   - "database does not exist": el usuario debe tener permiso CREATEDB.
   - Vendor inconsistente: ejecuta go mod vendor para regenerarlo.

8) DESARROLLO CON DOCKER (alternativa a go run)
   make dev          # levanta API + BD con hot-reload (Air)
   make dev-down     # apaga el stack

9) DEBUG CON BREAKPOINTS (VS Code)
   make debug        # levanta stack con Delve en puerto 40000
   # En VS Code: Run & Debug → "Go: Docker (Delve remoto)"
   make debug-down   # apaga y restaura dev

10) URL LOCAL
    http://localhost:8081/api/v1/invoices

