docker compose -f docker-compose-test.yaml up --abort-on-container-exit --build  
docker compose -f docker-compose-test.yaml -v down
cat ./report.json | go-test-report