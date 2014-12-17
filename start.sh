cd ./provisioner && go build -v . && mv provisioner ../bin/provisioner && cd .. && ./bin/provisioner & echo $! > ./bin/prov.pid
