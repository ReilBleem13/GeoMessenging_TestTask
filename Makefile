include .env
export

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down -v

docker-stop-one:
	docker compose stop $(name)

docker-restart-one: docker-stop-one docker-up
