.PHONY: dev dev-backend dev-frontend build lint install clean

dev:
	@make -j2 dev-backend dev-frontend

dev-backend:
	cd backend && npm run dev

dev-frontend:
	cd frontend && npm run dev

build:
	cd frontend && npm run build
	cd backend && npm run build

lint:
	cd backend && npm run lint
	cd frontend && npm run lint

install:
	cd backend && NODE_EXTRA_CA_CERTS=/etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem npm install
	cd frontend && NODE_EXTRA_CA_CERTS=/etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem npm install

clean:
	rm -rf backend/dist frontend/dist
