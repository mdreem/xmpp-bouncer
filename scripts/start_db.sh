#!/bin/bash
docker run -e MYSQL_ROOT_PASSWORD=password -e MYSQL_DATABASE=database --rm -p 3306:3306 mysql:8
