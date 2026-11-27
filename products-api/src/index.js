const express = require('express');
const { db, DB_PATH } = require('./db');

const app = express();
const PORT = Number(process.env.PORT) || 8003;

app.use(express.json());

app.get('/health', (_req, res) => {
  const { total } = db.prepare('SELECT COUNT(*) AS total FROM products').get();
  res.json({
    service: 'products-api',
    language: 'JavaScript / Express',
    status: 'UP',
    database: DB_PATH,
    records: total,
  });
});

app.post('/products', (req, res) => {
  const { name, description = '', price, stock = 0, category = 'General' } = req.body || {};

  if (!name || typeof price !== 'number' || Number.isNaN(price) || price < 0) {
    return res.status(400).json({ error: 'Debe proporcionar name y un price numérico válido (>= 0)' });
  }
  if (typeof stock !== 'number' || !Number.isInteger(stock) || stock < 0) {
    return res.status(400).json({ error: 'El stock debe ser un número entero mayor o igual a 0' });
  }

  const info = db
    .prepare('INSERT INTO products (name, description, price, stock, category) VALUES (?, ?, ?, ?, ?)')
    .run(String(name).trim(), String(description).trim(), price, stock, String(category).trim());

  const created = db.prepare('SELECT * FROM products WHERE id = ?').get(info.lastInsertRowid);
  return res.status(201).json({ message: 'Producto registrado correctamente', data: created });
});

app.get('/products', (_req, res) => {
  const rows = db.prepare('SELECT * FROM products ORDER BY id ASC').all();
  res.json({ total: rows.length, data: rows });
});

app.get('/products/:id', (req, res) => {
  const id = Number(req.params.id);
  if (!Number.isInteger(id)) {
    return res.status(400).json({ error: 'ID de producto inválido' });
  }
  const row = db.prepare('SELECT * FROM products WHERE id = ?').get(id);
  if (!row) {
    return res.status(404).json({ error: `Producto con ID ${id} no encontrado` });
  }
  return res.json({ data: row });
});

app.put('/products/:id', (req, res) => {
  const id = Number(req.params.id);
  if (!Number.isInteger(id)) {
    return res.status(400).json({ error: 'ID de producto inválido' });
  }

  const existing = db.prepare('SELECT * FROM products WHERE id = ?').get(id);
  if (!existing) {
    return res.status(404).json({ error: `Producto con ID ${id} no encontrado` });
  }

  const { name, description, price, stock, category } = req.body || {};
  const nextName = name !== undefined ? String(name).trim() : existing.name;
  const nextDesc = description !== undefined ? String(description).trim() : existing.description;
  const nextPrice = price !== undefined ? Number(price) : existing.price;
  const nextStock = stock !== undefined ? Number(stock) : existing.stock;
  const nextCat = category !== undefined ? String(category).trim() : existing.category;

  if (!nextName || Number.isNaN(nextPrice) || nextPrice < 0 || !Number.isInteger(nextStock) || nextStock < 0) {
    return res.status(400).json({ error: 'Datos de actualización inválidos' });
  }

  db.prepare(
    `UPDATE products
     SET name = ?, description = ?, price = ?, stock = ?, category = ?, updated_at = datetime('now')
     WHERE id = ?`
  ).run(nextName, nextDesc, nextPrice, nextStock, nextCat, id);

  const updated = db.prepare('SELECT * FROM products WHERE id = ?').get(id);
  return res.json({ message: 'Producto actualizado correctamente', data: updated });
});

app.delete('/products/:id', (req, res) => {
  const id = Number(req.params.id);
  if (!Number.isInteger(id)) {
    return res.status(400).json({ error: 'ID de producto inválido' });
  }
  const info = db.prepare('DELETE FROM products WHERE id = ?').run(id);
  if (info.changes === 0) {
    return res.status(404).json({ error: `Producto con ID ${id} no encontrado` });
  }
  return res.json({ message: `Producto ${id} eliminado correctamente`, deleted_id: id });
});

app.listen(PORT, '0.0.0.0', () => {
  console.log(`Products API (Node/Express) escuchando en :${PORT} con SQLite en ${DB_PATH}`);
});
