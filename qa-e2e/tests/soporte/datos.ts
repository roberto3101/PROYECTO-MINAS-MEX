import pkg from "pg";

const { Pool } = pkg;

export const CADENA_QA =
  process.env.CADENA_QA ?? "postgres://postgres:x@127.0.0.1:5433/mina_qa";

const grupo = new Pool({ connectionString: CADENA_QA, max: 4 });

export async function consultar<T = Record<string, unknown>>(
  sql: string,
  parametros: unknown[] = []
): Promise<T[]> {
  const resultado = await grupo.query(sql, parametros);
  return resultado.rows as T[];
}

export async function contar(sql: string, parametros: unknown[] = []): Promise<number> {
  const filas = await consultar<{ total: string }>(sql, parametros);
  return Number(filas[0]?.total ?? 0);
}

export async function comoEmpresa<T = Record<string, unknown>>(
  identificadorEmpresa: string,
  sql: string,
  parametros: unknown[] = []
): Promise<T[]> {
  const conexion = await grupo.connect();
  try {
    await conexion.query("BEGIN");
    await conexion.query("SELECT set_config('app.empresa_actual', $1, true)", [identificadorEmpresa]);
    await conexion.query("SET LOCAL ROLE aplicacion");
    const resultado = await conexion.query(sql, parametros);
    await conexion.query("COMMIT");
    return resultado.rows as T[];
  } catch (error) {
    await conexion.query("ROLLBACK");
    throw error;
  } finally {
    conexion.release();
  }
}

export async function cerrarConexiones(): Promise<void> {
  await grupo.end();
}
