@echo off
setlocal
title Plataforma Minera - arranque local
set "RAIZ=%~dp0"
set "PG=C:\Program Files\PostgreSQL\17\bin"
set "DATA=%RAIZ%.devdb\data"

echo ====================================================
echo  Plataforma Minera - arranque local
echo ====================================================
echo.
echo [1/2] Verificando PostgreSQL en el puerto 5433...
"%PG%\pg_isready.exe" -h 127.0.0.1 -p 5433 >nul 2>&1
if errorlevel 1 (
  echo      No estaba arriba: levantando el cluster...
  "%PG%\pg_ctl.exe" -D "%DATA%" -o "-p 5433" -l "%RAIZ%.devdb\arranque.log" start
  timeout /t 3 >nul
) else (
  echo      Ya estaba arriba.
)
"%PG%\pg_isready.exe" -h 127.0.0.1 -p 5433
echo.
echo [2/2] Iniciando el servidor Go en http://localhost:8080
echo      (Ctrl+C para detener)
echo.
cd /d "%RAIZ%backend"
set "CADENA_POSTGRES=postgres://postgres:x@127.0.0.1:5433/mina"
set "SECRETO_TOKEN=secreto-local"
set "DIRECCION=:8080"
set "DIRECTORIO_FRONTEND=%RAIZ%frontend"
set "DIRECTORIO_ARCHIVOS=%RAIZ%datos"
go run ./cmd/servidor
