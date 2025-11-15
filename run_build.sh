#!/bin/bash

set -e 

# use sudo here so that the user can enter their password and the second sudo will run without prompting the user for their password
sudo echo "starting build process"

cd frontend
npm run build

cd ../backend
go build

cd ..
sudo systemctl restart potbot

echo "successfully built frontend and backend. Changes should now be live."
