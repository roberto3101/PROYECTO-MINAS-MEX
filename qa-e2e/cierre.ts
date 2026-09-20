import { cerrarConexiones } from "./tests/soporte/datos";

export default async function cerrarTodo(): Promise<void> {
  await cerrarConexiones();
}
