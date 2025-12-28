#!/bin/bash
# Generate self-signed SSL certificate for local testing
# Usage: ./scripts/generate-self-signed-cert.sh [domain]

set -e

DOMAIN=${1:-localhost}
CERT_DIR="./ssl"
DAYS=365

echo "🔐 Generating self-signed certificate for: $DOMAIN"

# Create ssl directory if it doesn't exist
mkdir -p "$CERT_DIR"

# Generate private key
openssl genrsa -out "$CERT_DIR/server.key" 2048

# Generate certificate signing request (CSR)
openssl req -new -key "$CERT_DIR/server.key" -out "$CERT_DIR/server.csr" -subj "/C=US/ST=State/L=City/O=Organization/OU=IT/CN=$DOMAIN"

# Generate self-signed certificate
openssl x509 -req -days $DAYS -in "$CERT_DIR/server.csr" -signkey "$CERT_DIR/server.key" -out "$CERT_DIR/server.crt"

# Set permissions
chmod 600 "$CERT_DIR/server.key"
chmod 644 "$CERT_DIR/server.crt"

# Clean up CSR
rm "$CERT_DIR/server.csr"

echo "✅ Certificate generated successfully!"
echo ""
echo "📁 Files created:"
echo "   Certificate: $CERT_DIR/server.crt"
echo "   Private Key: $CERT_DIR/server.key"
echo ""
echo "🚀 To use with apprun:"
echo "   export SSL_CERT_FILE=$PWD/$CERT_DIR/server.crt"
echo "   export SSL_KEY_FILE=$PWD/$CERT_DIR/server.key"
echo "   go run core/cmd/server/main.go"
echo ""
echo "⚠️  This is a self-signed certificate. Browsers will show a warning."
echo "   For production, use Let's Encrypt or a proper CA-signed certificate."
