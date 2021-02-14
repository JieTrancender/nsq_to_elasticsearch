set -ex

# containerId=$(docker ps |grep etcd |awk '{print $1}')
# /usr/bin/docker exec -it ${containerId} /bin/sh -c 'etcdctl user add root --new-user-password 123456'
# /usr/bin/docker exec -it ${containerId} /bin/sh -c 'etcdctl role add root'
# /usr/bin/docker exec -it ${containerId} /bin/sh -c 'etcdctl user grant-role root root'
# /usr/bin/docker exec -it ${containerId} /bin/sh -c 'etcdctl auth enable'

# add root user
curl -L http://localhost:2379/v3/auth/user/add \
  -X POST -d '{"name": "root", "password": "root"}'

# add root role
curl -L http://localhost:2379/v3/auth/role/add \
  -X POST -d '{"name": "root"}'

# grant root role to root user
curl -L http://localhost:2379/v3/auth/user/grant \
  -X POST -d '{"user": "root", "role": "root"}'

# enable auth
curl -L http://localhost:2379/v3/auth/enable -X POST -d '{}'

# get token
curl -L http://localhost:2379/v3/auth/authenticate \
  -X POST -d '{"name": "root", "password": "root"}'

curl -ig 'localhost:2379/ping'
