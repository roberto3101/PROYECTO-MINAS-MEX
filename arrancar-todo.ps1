param(
  [int]$Puerto = 5433,
  [string]$Base = 'mina',
  [string]$Data = '',
  [switch]$SoloPreparar,
  [switch]$Reiniciar
)

$ErrorActionPreference = 'Stop'
$Raiz = Split-Path $MyInvocation.MyCommand.Path -Parent
if ([string]::IsNullOrWhiteSpace($Data)) { $Data = Join-Path $Raiz '.devdb\data' }
$Log = Join-Path (Split-Path $Data -Parent) 'arranque.log'
$env:PGPASSWORD = 'x'

function Fallo($mensaje) { Write-Host ''; Write-Host "  ERROR: $mensaje" -ForegroundColor Red; exit 1 }

Write-Host '===================================================='
Write-Host ' Plataforma Minera - instalacion y arranque (un clic)'
Write-Host '===================================================='

$pgBin = $null
foreach ($version in 17, 16, 15, 14) {
  $candidato = "C:\Program Files\PostgreSQL\$version\bin"
  if (Test-Path (Join-Path $candidato 'pg_ctl.exe')) { $pgBin = $candidato; break }
}
if (-not $pgBin) {
  $enPath = (Get-Command psql -ErrorAction SilentlyContinue).Source
  if ($enPath) { $pgBin = Split-Path $enPath -Parent }
}
if (-not $pgBin) { Fallo 'No encontre PostgreSQL. Instala PostgreSQL 17 o agrega su carpeta bin al PATH.' }
$env:PATH = "$pgBin;$env:PATH"
$initdb = Join-Path $pgBin 'initdb.exe'
$pgctl = Join-Path $pgBin 'pg_ctl.exe'
$isready = Join-Path $pgBin 'pg_isready.exe'
$psqlExe = Join-Path $pgBin 'psql.exe'
Write-Host "PostgreSQL: $pgBin"

if ($Reiniciar -and (Test-Path (Join-Path $Data 'PG_VERSION'))) {
  Write-Host '[0/6] Reinicio solicitado: deteniendo y borrando el cluster anterior...'
  try { Start-Process -FilePath $pgctl -ArgumentList @('-D', $Data, 'stop', '-m', 'immediate') -NoNewWindow -Wait } catch {}
  Start-Sleep -Seconds 1
  Remove-Item -Recurse -Force $Data
}

if (-not (Test-Path (Join-Path $Data 'PG_VERSION'))) {
  Write-Host '[1/6] Creando el cluster PostgreSQL local (initdb)...'
  New-Item -ItemType Directory -Force -Path (Split-Path $Data -Parent) | Out-Null
  & $initdb -D $Data -U postgres -A trust -E UTF8 --locale=C 1>$null
  if ($LASTEXITCODE -ne 0) { Fallo 'initdb no pudo crear el cluster.' }
} else {
  Write-Host '[1/6] Cluster ya existente.'
}

& $isready -h 127.0.0.1 -p $Puerto 1>$null 2>$null
if ($LASTEXITCODE -ne 0) {
  Write-Host "[2/6] Levantando PostgreSQL en el puerto $Puerto..."
  Start-Process -FilePath $pgctl -ArgumentList @('-D', $Data, '-o', "-p$Puerto", '-l', $Log, 'start') -NoNewWindow
  for ($i = 0; $i -lt 40; $i++) {
    Start-Sleep -Milliseconds 500
    & $isready -h 127.0.0.1 -p $Puerto 1>$null 2>$null
    if ($LASTEXITCODE -eq 0) { break }
  }
} else {
  Write-Host '[2/6] PostgreSQL ya estaba arriba.'
}
& $isready -h 127.0.0.1 -p $Puerto 1>$null 2>$null
if ($LASTEXITCODE -ne 0) { Fallo "PostgreSQL no acepto conexiones en el puerto $Puerto." }

$existe = ("$(& $psqlExe -U postgres -h 127.0.0.1 -p $Puerto -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname='$Base'")").Trim()
if ($existe -ne '1') {
  Write-Host "[3/6] Creando la base de datos '$Base'..."
  & $psqlExe -U postgres -h 127.0.0.1 -p $Puerto -d postgres -c "CREATE DATABASE $Base" 1>$null
  if ($LASTEXITCODE -ne 0) { Fallo "No pude crear la base '$Base'." }
} else {
  Write-Host "[3/6] La base '$Base' ya existe."
}

$tieneEsquema = ("$(& $psqlExe -U postgres -h 127.0.0.1 -p $Puerto -d $Base -tAc "SELECT to_regclass('gobierno.empresa')")").Trim()
if ([string]::IsNullOrWhiteSpace($tieneEsquema)) {
  Write-Host '[4/6] Cargando esquema, vistas y datos semilla desde los .sql...'
  $ejecutar = Join-Path $Raiz 'infraestructura\base-de-datos\ejecutar.ps1'
  & powershell -NoProfile -ExecutionPolicy Bypass -File $ejecutar -DbHost 127.0.0.1 -Puerto $Puerto -BaseDatos $Base -Usuario postgres -Clave x
  if ($LASTEXITCODE -ne 0) { Fallo 'Fallo la carga de los .sql.' }
} else {
  Write-Host '[4/6] El esquema ya estaba cargado (no se re-aplica).'
}

$exe = Join-Path (Split-Path $Data -Parent) 'servidor.exe'
Write-Host '[5/6] Compilando el servidor Go...'
Push-Location (Join-Path $Raiz 'backend')
& go build -o $exe ./cmd/servidor
$codigoBuild = $LASTEXITCODE
Pop-Location
if ($codigoBuild -ne 0) { Fallo 'Fallo la compilacion del backend. Verifica que Go este instalado.' }

if ($SoloPreparar) {
  Write-Host '[6/6] Preparacion completa (no se arranca el servidor por -SoloPreparar).'
  Write-Host ''
  Write-Host "Base lista en 127.0.0.1:$Puerto/$Base y binario en $exe"
  exit 0
}

$env:CADENA_POSTGRES = "postgres://postgres:x@127.0.0.1:$Puerto/$Base"
$env:SECRETO_TOKEN = 'secreto-local'
$env:DIRECCION = ':8080'
$env:DIRECTORIO_FRONTEND = (Join-Path $Raiz 'frontend')
$env:DIRECTORIO_ARCHIVOS = (Join-Path $Raiz 'datos')
Write-Host '[6/6] Servidor listo en http://localhost:8080'
Write-Host ''
Write-Host '  Entra con:  empresa MIN  /  usuario admin.mina  /  clave Mina#2026'
Write-Host '  Superadmin: usuario plataforma  /  clave Plataforma#2026'
Write-Host '  (Ctrl+C en esta ventana para detener el servidor)'
Write-Host ''
Start-Process 'http://localhost:8080'
& $exe
