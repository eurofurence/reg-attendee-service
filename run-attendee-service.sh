#! /bin/bash

STARTTIME=$(date '+%Y-%m-%d_%H-%M-%S')

echo "Writing log to ~/work/logs/attendee-service.$STARTTIME.log"

cd ~/work/attendee-service

./attendee-service -config config.yaml -migrate-database &> ~/work/logs/attendee-service.$STARTTIME.log

