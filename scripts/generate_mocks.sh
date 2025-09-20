#!/bin/bash

# Generate mocks for all interfaces

echo "Generating mocks..."

# Create mocks directory if it doesn't exist
mkdir -p tests/helpers/mocks

# Generate repository mocks
mockgen -source=src/repositories/interfaces.go -destination=tests/helpers/mocks/mock_repositories.go -package=mocks

# Generate service mocks
mockgen -source=src/services/interfaces.go -destination=tests/helpers/mocks/mock_services.go -package=mocks

# Generate client mocks
mockgen -source=src/clients/interfaces.go -destination=tests/helpers/mocks/mock_clients.go -package=mocks

echo "Mocks generated successfully!"