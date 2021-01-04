## go-docker
>a docker implement for golang

## using
```bash
# complie
make and cd ./bin

# run go-docker using busybox, busybox place in /root/busybox.tar
./go-docker run -ti --name test busybox sh

# run go-docker as daemon
./go-docker run -d --name test busybox sh

# mount files
./go-docker run -d -v /root/test:/test --name test busybox.sh

# enter go-docker
./go-docker exec test sh

# look logs
./go-docker logs test

# list go-docker ps
./go-docker ps 

# stop go-docker
./go-docker stop test

# remove go-docker
./go-docker rm test
```

## knowledge
### namespace
- uts: isolate host
- pid: isolate process pid
- user: isolate user
- network: isolate network
- mount: isolate mount
- ipc: isolate system V IPC and POSIX message queue

### cgroups
- cgroup: manage pid
- subsystem: mange system source
- hierarchy: manage a cgroup as tree type 