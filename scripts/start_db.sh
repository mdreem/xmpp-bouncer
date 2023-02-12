#!/bin/bash
docker run -e MYSQL_ROOT_PASSWORD=password -e MYSQL_DATABASE=xmpp-bouncer --rm -p 3306:3306 --name xmpp-mysql mysql:8
