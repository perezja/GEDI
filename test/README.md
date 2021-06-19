# How to initiate a Chord network

## 1. Start Up the Satellite

Start a docker container the satellite node

```
sudo docker run -it --name sat --mount type=bind,source=/home/apollodorus/go/src/github.com/gedilabs,target=/go/src/github.com/gedilabs -w /go/src/github.com/gedilabs gedilabs/satellite:v1
```

Run postgresql server and start a new pseudoterminal into the running satellite container in a separate window to use the postgres terminal

```
service postgresql start
sudo docker exec -it sat /bin/bash
su postgres
```

Make keys before running a `test_satellite.go`

```
ssh-keygen
```

## 2. Start Up Chord Nodes

```
sudo docker run -it --name node --mount type=bind,source=/home/apollodorus/go/src/github.com/gedilabs,target=/go/src/github.com/gedilabs -w /go/src/github.com/gedilabs gedilabs/base:v1
```


