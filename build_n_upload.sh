#! /usr/bin/env bash
set -e
go test
GOOS=linux go build -o countmyreps-new
scp -r {countmyreps-new,web,go_templates} $MYIP:~/countmyreps/.
ssh -t $MYIP "sudo service countmyreps stop && mv ~/countmyreps/countmyreps-new ~/countmyreps/countmyreps-linux && sudo service countmyreps start"
