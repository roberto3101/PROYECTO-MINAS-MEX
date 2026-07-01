@echo off
title Plataforma Minera - Arranque de un clic
chcp 65001 >nul
powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0arrancar-todo.ps1" %*
echo.
echo El servidor se detuvo o hubo un error. Presiona una tecla para cerrar.
pause >nul
