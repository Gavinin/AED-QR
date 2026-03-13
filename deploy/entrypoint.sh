#!/bin/sh

# Entrypoint script to handle initialization

# 1. Config Check
if [ ! -f /app/config/config.yml ]; then
    echo "Config file not found. Copying default..."
    cp /app/example/config.yml /app/config/config.yml
fi

# 2. Database Initialization
if [ ! -f /app/db/index.db ]; then
    if [ -f /app/example/index.db ]; then
        echo "Database file not found. Copying from example..."
        cp /app/example/index.db /app/db/index.db
    else
        echo "Database file not found. It will be created by the backend."
        touch /app/db/index.db
    fi
fi

# 3. SSL Certificates
# If user provided certs in volume, use them. Otherwise generate self-signed for testing.
if [ ! -f /etc/nginx/ssl/cert.pem ]; then
    echo "SSL certificates not found. Generating self-signed..."
    mkdir -p /etc/nginx/ssl
    openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
        -keyout /etc/nginx/ssl/key.pem \
        -out /etc/nginx/ssl/cert.pem \
        -subj "/C=CN/ST=State/L=City/O=Organization/CN=localhost"
fi

# 4. Start Backend (Background)
echo "Starting Backend..."
./backend &

# 5. Start Nginx (Foreground)
echo "Starting Nginx..."
nginx -g "daemon off;"
