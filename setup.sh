#!/bin/bash

# Create necessary directories
echo "Creating directories..."
mkdir -p config db log ssl

# Copy config if not exists
if [ ! -f config/config.yml ]; then
    echo "Copying config.yml from example..."
    cp example/config.yml config/config.yml
else
    echo "config/config.yml already exists, skipping."
fi

# Copy database if not exists
if [ ! -f db/index.db ]; then
    if [ -f example/index.db ]; then
        echo "Copying index.db from example..."
        cp example/index.db db/index.db
    else
        echo "Creating empty index.db..."
        touch db/index.db
    fi
else
    echo "db/index.db already exists, skipping."
fi

echo "Setup complete. You can now run 'docker-compose up -d --build'"
