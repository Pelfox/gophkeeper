# gophkeeper

Secure secrets manager written in Go.

Gophkeeper supports storing multiple types of vault items:

- Login credentials
- Text notes
- Binary files
- Bank card details

## Architecture

Gophkeeper is a client-server application. The client is a CLI application
available for macOS, Linux, and Windows. Communication between the client and
server is done via HTTP/2, with TLSv1.3 security.

All encryption is performed on the client side. After vault creation, client
generates a unique 256-bit (32 bytes) master key. The master key is encrypted
before being stored in the database. Vault items are encrypted individually
using the user's master key.

User passwords are hashed using Argon2id. Vault items are encrypted with
ChaCha20-Poly1305, and each item uses its own unique nonce.

## Getting started

To download a prebuilt binary for your platform, open this repository's
"Releases" tab. Binaries for all supported platforms are available there.
