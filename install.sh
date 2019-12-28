#! /bin/bash

set -o errexit

if [[ "$RUNTIME_USER" == "" ]]; then
  echo "RUNTIME_USER not set, bailing out. Please run setup.sh first."
  exit 1
fi

mkdir -p tmp
cp attendee-service tmp/
cp config.yaml tmp/
cp run-attendee-service.sh tmp/

chgrp $RUNTIME_USER tmp/*
chmod 640 tmp/config.yaml
chmod 750 tmp/attendee-service
chmod 750 tmp/run-attendee-service.sh
mv tmp/attendee-service /home/$RUNTIME_USER/work/attendee-service/
mv tmp/config.yaml /home/$RUNTIME_USER/work/attendee-service/
mv tmp/run-attendee-service.sh /home/$RUNTIME_USER/work/

