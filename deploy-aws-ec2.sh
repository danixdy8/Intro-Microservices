#!/usr/bin/env bash
# ==============================================================================
# SCRIPT DE DESPLIEGUE EN AWS EC2 (Ubuntu 22.04 / 24.04 LTS)
# Optimizado para instancias de bajo costo / Capa Gratuita (t2.micro / t3.micro)
# ==============================================================================
set -e

echo "=== 1. Creando Swapfile de 2GB (Evita caidas por OOM en instancias de 1GB RAM) ==="
if [ ! -f /swapfile ]; then
    sudo fallocate -l 2G /swapfile || sudo dd if=/dev/zero of=/swapfile bs=1M count=2048
    sudo chmod 600 /swapfile
    sudo mkswap /swapfile
    sudo swapon /swapfile
    echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab
    echo "Swapfile creado con éxito."
else
    echo "Swapfile ya existente."
fi

echo "=== 2. Actualizando paquetes del sistema ==="
sudo apt-get update -y
sudo apt-get install -y ca-certificates curl gnupg lsb-release

echo "=== 3. Instalando Docker y Docker Compose ==="
if ! command -v docker &> /dev/null; then
    curl -fsSL https://get.docker.com -o get-docker.sh
    sudo sh get-docker.sh
    rm get-docker.sh
    sudo usermod -aG docker $USER
    echo "Docker instalado correctamente."
else
    echo "Docker ya se encuentra instalado."
fi

echo "=== 4. Levantando los 4 microservicios con Docker Compose ==="
sudo docker compose up -d --build

echo "=== 5. Estado de los contenedores ==="
sudo docker compose ps

echo "=============================================================================="
echo "¡Microservicios levantados correctamente!"
echo "Recuerda abrir los siguientes puertos en el Security Group de tu EC2 en AWS:"
echo " - 8000 (Users API - Python)"
echo " - 8001 (Auth API - Java)"
echo " - 8002 (Orders API - Go)"
echo " - 8003 (Products API - JavaScript)"
echo "=============================================================================="
