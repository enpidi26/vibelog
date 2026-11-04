#!/bin/bash
# Script untuk deploy VibeLog ke Coolify
# Jalankan di VPS: bash deploy-coolify.sh

set -e

echo "🚀 Deploying VibeLog to Coolify..."

# Clone repository ke temporary directory
TEMP_DIR="/tmp/vibelog-deploy-$(date +%s)"
echo "📦 Cloning repository..."
git clone -b feat/animasi-ui https://github.com/enpidi26/vibelog.git "$TEMP_DIR"
cd "$TEMP_DIR"

# Generate JWT Secret
JWT_SECRET=$(openssl rand -base64 48)
echo "🔐 Generated JWT_SECRET"

# Create .env file
cat > .env << EOF
APP_ENV=production
HOST=0.0.0.0
PORT=8080
DATABASE_PATH=/data/vibelog.db
JWT_SECRET=$JWT_SECRET
JWT_EXPIRY_HOURS=24
FRONTEND_URL=http://201.79.105.137
EOF

echo "✅ Created .env file"

# Build and deploy with docker-compose
echo "🐳 Building and starting containers..."
docker compose up -d --build

echo ""
echo "✅ Deployment complete!"
echo ""
echo "📋 Access your app at: http://201.79.105.137"
echo ""
echo "🔑 Your JWT_SECRET (save this securely):"
echo "$JWT_SECRET"
echo ""
echo "📊 Check status: docker compose ps"
echo "📝 View logs: docker compose logs -f"
echo ""
