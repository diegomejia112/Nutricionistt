#!/bin/bash
# Desarrollo: Go backend + Vite. Mata procesos previos, auto-reinicia si algo muere.

cleanup() {
  echo "Deteniendo..."
  kill $(jobs -p) 2>/dev/null
  wait 2>/dev/null
  fuser -k 3847/tcp 5173/tcp 2>/dev/null || true
  exit 0
}
trap cleanup EXIT INT TERM

# Limpiar puertos previos
fuser -k 3847/tcp 5173/tcp 2>/dev/null || true
sleep 0.5

echo "==> Backend Go en :3847"
(while true; do
  go run . 2>&1
  echo "[backend] reiniciando en 2s..."
  sleep 2
done) &

echo "==> Frontend Vite en :5173"
(cd frontend && while true; do
  npm run dev 2>&1
  echo "[frontend] reiniciando en 2s..."
  sleep 2
done) &

echo ""
echo "Nutricionist en desarrollo:"
echo "  http://localhost:5173"
echo ""
echo "Ctrl+C para detener todo."

wait
