.PHONY: dev dev-backend dev-frontend build lint install clean

# Fedora / RHEL hosts lack the Google Trust Services intermediate CA that
# Anthropic's API uses.  Set NODE_EXTRA_CA_CERTS only when the bundle exists
# so the fix is a no-op on macOS, Ubuntu, Alpine, etc.
CA_BUNDLE := /etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem
CA_ENV    := $(if $(wildcard $(CA_BUNDLE)),NODE_EXTRA_CA_CERTS=$(CA_BUNDLE),)

dev:
	@make -j2 dev-backend dev-frontend

dev-backend:
	cd backend && $(CA_ENV) npm run dev

dev-frontend:
	cd frontend && npm run dev

build:
	cd frontend && npm run build
	cd backend && npm run build

lint:
	cd backend && npm run lint
	cd frontend && npm run lint

install:
	cd backend && npm install
	cd frontend && npm install

clean:
	rm -rf backend/dist frontend/dist
