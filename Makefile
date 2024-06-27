PWD = $(shell pwd)

image: image-backend image-frontend

image-backend:
	docker build -f backend/productsearch/build/Dockerfile -t productsearch:latest ./backend/productsearch

image-frontend:
	docker build -f frontend/psf/Dockerfile -t psf:latest ./frontend/psf

up:
	docker-compose up -d

stop:
	docker-compose stop

down:
	docker-compose down
