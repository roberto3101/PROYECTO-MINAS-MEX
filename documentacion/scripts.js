/* ============================================================
   Scripts compartidos · zoom de diagramas + tooltips SQL
   (Mermaid se inicializa en cada página con su <script type=module>)
   ============================================================ */
window.addEventListener('load', function () {
  // --- Zoom / pantalla completa de diagramas ---
  document.querySelectorAll('.dgrm').forEach(function (d) {
    var s = d.querySelector('.dgrm-scale'), k = 1, ap = function () { if (s) s.style.transform = 'scale(' + k + ')'; };
    var bo = d.querySelector('[data-z=out]'), bi = d.querySelector('[data-z=in]'),
        br = d.querySelector('[data-z=reset]'), bf = d.querySelector('[data-z=full]');
    if (bo) bo.addEventListener('click', function () { k = Math.max(k - 0.2, 0.4); ap(); });
    if (bi) bi.addEventListener('click', function () { k = Math.min(k + 0.2, 3); ap(); });
    if (br) br.addEventListener('click', function () { k = 1; ap(); });
    if (bf) bf.addEventListener('click', function () { d.classList.toggle('full'); });
  });
  document.addEventListener('keydown', function (e) {
    if (e.key === 'Escape') document.querySelectorAll('.dgrm.full').forEach(function (d) { d.classList.remove('full'); });
  });
});

/* --- Tooltips automáticos por línea de SQL (PostgreSQL) --- */
(function () {
  function textoDe(html) { var d = document.createElement('div'); d.innerHTML = html; return d.textContent; }
  function entreParens(t) { var m = t.match(/\(([^)]*)\)/); return m ? m[1].trim() : ''; }

  function explicarTipoInline(t) {
    var col = (t.match(/^([a-z_]+)\s+/) || [])[1];
    if (!col) return '';
    var ob = /NOT NULL/.test(t) ? ' Obligatorio.' : ' Opcional.';
    if (/NUMERIC\(/.test(t)) return 'Valor numérico exacto «' + col + '».' + ob;
    if (/\bSMALLINT\b/.test(t)) return 'Entero corto «' + col + '».' + ob;
    if (/\bINTEGER\b/.test(t)) return 'Número entero «' + col + '».' + ob;
    if (/TIMESTAMPTZ/.test(t)) return 'Fecha y hora con zona «' + col + '».' + ob;
    if (/\bDATE\b/.test(t)) return 'Fecha de negocio «' + col + '».' + ob;
    if (/\bBOOLEAN\b/.test(t)) return 'Sí/No (booleano) «' + col + '».' + ob;
    if (/\bTEXT\b/.test(t)) return 'Texto «' + col + '».' + ob;
    if (/\bUUID\b/.test(t)) return 'Identificador «' + col + '».' + ob;
    return '';
  }

  function explicar(raw) {
    var t = raw.trim(), m;
    if (t === '') return '';
    if (/^--/.test(t)) return 'Comentario.';
    if (m = t.match(/^CREATE TABLE\s+([a-z_\.]+)/i)) return 'Crea la tabla «' + m[1] + '» (entidad del dominio).';
    if (/^\)\s*;?\s*$/.test(t)) return 'Cierra la definición de la tabla.';
    if (/PRIMARY KEY/.test(t) && /gen_random_uuid/.test(t)) return 'Clave primaria: UUID generado automáticamente.';
    if (/^id_empresa\b/.test(t)) return 'Tenant (multiempresa): empresa dueña del registro. Llave foránea a «empresa».';
    if (/^id_obra\b/.test(t)) return 'Obra (labor minera) a la que pertenece el registro. Llave foránea.';
    if (/^id_mina\b/.test(t)) return 'Mina del registro (llave foránea).';
    if (/^id_equipo\b/.test(t)) return 'Equipo usado (llave foránea).';
    if (/^id_[a-z_]+\b/.test(t)) { var c = t.match(/^(id_[a-z_]+)/)[1]; return 'Referencia (llave foránea): ' + c + '.'; }
    if (/^creado_en\b/.test(t)) return 'Auditoría: fecha y hora de creación.';
    if (/^actualizado_en\b/.test(t)) return 'Auditoría: fecha y hora de la última modificación.';
    if (/^eliminado_en\b/.test(t)) return 'Borrado lógico: con valor = dado de baja; NULL = activo.';
    if (/^(creado|actualizado|eliminado)_por_usuario_id\b/.test(t)) return 'Auditoría: usuario responsable de la acción.';
    if (/^estado\s+TEXT/.test(t)) { var d = t.match(/DEFAULT\s+'([^']+)'/); return 'Estado / ciclo de vida' + (d ? (' (por defecto «' + d[1] + '»)') : '') + '.'; }
    if (m = t.match(/^CONSTRAINT\s+ck_[a-z_]+\s+CHECK\s*\(\s*([a-z_]+)\s+IN\s*\(([^)]*)\)/i)) return 'Validación: «' + m[1] + '» solo puede ser: ' + m[2].replace(/'/g, '').trim() + '.';
    if (m = t.match(/^CONSTRAINT\s+ck_[a-z_]+\s+CHECK\s*\(([a-z_]+)\s*>=\s*([a-z_]+)\)/i)) return 'Coherencia: «' + m[1] + '» ≥ «' + m[2] + '».';
    if (/^CONSTRAINT\s+ck_[a-z_]+\s+CHECK/i.test(t)) return 'Validación (CHECK): ' + entreParens(t) + '.';
    if (m = t.match(/ADD CONSTRAINT\s+fk_[a-z_]+\s+FOREIGN KEY\s*\(([a-z_]+)\)\s+REFERENCES\s+([a-z_\.]+)/i)) return 'Llave foránea: «' + m[1] + '» debe existir en «' + m[2] + '».';
    if (/^(CONSTRAINT\s+fk_|.*FOREIGN KEY)/i.test(t)) return 'Llave foránea (integridad referencial).';
    if (/^ALTER TABLE/i.test(t)) return 'Restricción declarada aparte del CREATE.';
    if (m = t.match(/^CREATE\s+(UNIQUE\s+)?INDEX\s+([a-z_]+)/i)) { var u = /UNIQUE/i.test(t), p = /WHERE/i.test(t); return (u ? 'Índice ÚNICO' : 'Índice') + ' «' + m[2] + '»' + (p ? ' parcial (solo activos: permite re-alta tras baja lógica).' : (u ? ' — no permite duplicados.' : ' — acelera búsquedas.')); }
    if (/^COMMENT ON/i.test(t)) return 'Documentación de la tabla.';
    var ti = explicarTipoInline(t);
    if (ti) return ti;
    return 'Continúa la sentencia de la línea anterior.';
  }

  document.querySelectorAll('pre.sql').forEach(function (pre) {
    var lineas = pre.innerHTML.split('\n');
    pre.innerHTML = lineas.map(function (lineaHTML) {
      var tip = explicar(textoDe(lineaHTML));
      if (!tip) return lineaHTML;
      return '<span class="linea-sql" data-tip="' + tip.replace(/"/g, '&quot;') + '">' + lineaHTML + '</span>';
    }).join('\n');
  });
})();
