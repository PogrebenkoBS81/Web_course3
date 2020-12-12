Run Redis:
docker run -p 6379:6379 -e ALLOW_EMPTY_PASSWORD=yes bitnami/redis:latest

Build image:
docker build <your path here>\lab4-5\activity_api -t activity_api

Run:
docker run -it --rm -p 9332:9332 activity_api