# ============================================================
# ejecutar.ps1 - Crea/recarga el esquema completo + seed + tests
# ============================================================
# Uso:
#   .\ejecutar.ps1                 # usa el cluster de pruebas (127.0.0.1:5433, BD 'mina')
#   .\ejecutar.ps1 -Puerto 5432 -BaseDatos mina_dev -Usuario postgres
#
# Requiere psql en el PATH (o ajustar $PsqlExe). No incluye el seed/tests
# en produccion: pasar -SoloEsquema para cargar solo el DDL y las vistas.

param(
  [string]$DbHost = '127.0.0.1',
  [int]$Puerto = 5433,
  [string]$BaseDatos = 'mina',
  [string]$Usuario = 'postgres',
  [string]$Clave = 'x',
  [switch]$SoloEsquema
)

$ErrorActionPreference = 'Stop'
$env:PGPASSWORD = $Clave
$PsqlExe = (Get-Command psql -ErrorAction SilentlyContinue).Source
if (-not $PsqlExe) { $PsqlExe = 'C:\Program Files\PostgreSQL\17\bin\psql.exe' }
$Raiz = Split-Path $MyInvocation.MyCommand.Path -Parent

$esquema = @(
  'esquema\00_reset_y_esquemas.sql','esquema\01_gobierno.sql','esquema\02_catalogos.sql',
  'esquema\03_produccion.sql','esquema\04_planeacion.sql','esquema\05_reconciliacion.sql',
  'esquema\06_costos.sql','esquema\07_explosivo.sql','esquema\08_beneficio.sql',
  'esquema\09_estandares.sql','esquema\10_inversiones.sql','esquema\11_indices.sql',
  'esquema\12_integridad_tenant.sql','esquema\50_vistas.sql','esquema\60_seguridad_rls.sql'
)
$datos  = @('semilla\20_seed.sql')
$tests  = @('pruebas\30_tests.sql','pruebas\40_escenario_usuarios.sql')

$lista = $esquema
if (-not $SoloEsquema) { $lista += $datos + $tests }

$conn = @('-U',$Usuario,'-h',$DbHost,'-p',"$Puerto",'-d',$BaseDatos,'-v','ON_ERROR_STOP=1','-q')
foreach ($rel in $lista) {
  $f = Join-Path $Raiz $rel
  & $PsqlExe @conn -f $f
  if ($LASTEXITCODE -ne 0) { Write-Error "FALLO en $rel (exit $LASTEXITCODE)"; exit 1 }
  Write-Host ("ok  " + $rel)
}
Write-Host "`nEsquema cargado correctamente en $DbHost`:$Puerto/$BaseDatos"
