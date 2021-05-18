#!/bin/sh

docker run --name mariadbtest -e MYSQL_DATABASE=photoview_test -e MYSQL_USER=photoview -e MYSQL_PASSWORD=photosecret -e MYSQL_RANDOM_ROOT_PASSWORD=yes -p 3306:3306 -d docker.io/library/mariadb:10.3
