DOCKER_COMPOSE = docker-compose run --rm --remove-orphans terraform

help:
	@echo "commands:"
	@echo "make init: terraform init"
	@echo "make plan: terraform plan"
	@echo "make show: terraform show"
	@echo "make apply: terraform apply"
	@echo "make destroy: terraform destroy"
	@echo "make fmt: terraform fmt"
	@echo "make console: terraform console"

init:
	$(DOCKER_COMPOSE) init

plan:
	$(DOCKER_COMPOSE) plan

show:
	$(DOCKER_COMPOSE) show

apply:
	$(DOCKER_COMPOSE) apply

destroy:
	$(DOCKER_COMPOSE) destroy

fmt:
	$(DOCKER_COMPOSE) fmt -recursive

console:
	$(DOCKER_COMPOSE) console

output:
	$(DOCKER_COMPOSE) output

list:
	$(DOCKER_COMPOSE) state list