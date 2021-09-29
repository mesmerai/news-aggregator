#!/bin/bash

while read -r imageID
do 
    echo "Removing Image ID: ${imageID}"
    sudo docker rmi ${imageID}
done < <(sudo docker images | grep -v REPOSITORY | awk '{print $3}')

exit 0

