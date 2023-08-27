.PHONY: all test stop build down rebuild clean
all			:
		-docker-compose up -d noteservice db migrate
build		:
		-docker-compose build noteservice db migrate
down		:
		-docker-compose down
test		:
		-docker-compose up --force-recreate test
stop		:
		-docker-compose stop
retest		:
		-docker-compose build testuserapp testtaskapp
rebuild		:	clean all

clean		:	down
		-docker image rm test_container migrate_container db_container noteservice_container