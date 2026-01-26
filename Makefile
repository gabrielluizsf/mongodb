DOWN=docker compose -f docker-compose.yml down
DATABASE_NAME=atendi9
DATABASE_URI=mongodb://atendi9:docker@localhost:27017/

drop_database:
	@$(DOWN)

docker_system_prune:
	@docker system prune -a --volumes -f

prune_containers:
	@docker container prune -f

prune_volumes:
	@docker volume prune --all -f

test: drop_database prune_containers prune_volumes
	@docker compose down
	@docker compose -f docker-compose.yml up --build --force-recreate -d mongo
	@env MONGODB_URI=$(DATABASE_URI) DATABASE_NAME=$(DATABASE_NAME) go test ./... -v 
	@$(DOWN)