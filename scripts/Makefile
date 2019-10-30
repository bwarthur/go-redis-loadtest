.PHONEY: test
test:
	docker run --rm -it --network redis-ha_redis-ha redis redis-benchmark -q -n 100000 -h redis-envoy -p 6379 --dbnum 0 -t set,get,LPUSH,LPOP,INCR

.PHONEY: cli
cli:
	docker run --rm -it --network redis-ha_redis-ha redis redis-cli -h redis-envoy -p 6379

.PHONEY: kill-master
kill-master:
	docker kill redis-1

.PHONEY: config-replica
config-replica:
	 powershell -noprofile -command copy-item templates/envoy-replica.yaml envoy.yaml -force

	 powershell -noprofile -command copy-item templates/redis-master.conf redis-1.conf -force
	 powershell -noprofile -command copy-item templates/redis-slave.conf redis-2.conf -force
	 powershell -noprofile -command copy-item templates/redis-slave.conf redis-3.conf -force
	 
	 powershell -noprofile -command copy-item templates/sentinel-default.conf sentinel-1.conf -force
	 powershell -noprofile -command copy-item templates/sentinel-default.conf sentinel-2.conf -force
	 powershell -noprofile -command copy-item templates/sentinel-default.conf sentinel-3.conf -force

.PHONEY: run-replica
run-replica: config-replica
	docker-compose -f docker-compose-replica.yaml up

.PHONEY: config-cluster
config-cluster:
	 powershell -noprofile -command copy-item templates/envoy-cluster.yaml envoy.yaml -force

	 powershell -noprofile -command copy-item templates/redis-cluster.conf redis-1.conf -force
	 powershell -noprofile -command copy-item templates/redis-cluster.conf redis-2.conf -force
	 powershell -noprofile -command copy-item templates/redis-cluster.conf redis-3.conf -force

	 powershell -noprofile -command copy-item templates/redis-cluster.conf redis-w1.conf -force
	 powershell -noprofile -command copy-item templates/redis-cluster.conf redis-w2.conf -force
	 powershell -noprofile -command copy-item templates/redis-cluster.conf redis-w3.conf -force

.PHONEY: run-cluster
run-cluster: config-cluster
	docker-compose -f docker-compose-cluster.yaml up


.PHONEY: clean
clean:
	docker-compose -f docker-compose-cluster.yaml rm --force --stop
	docker-compose -f docker-compose-replica.yaml rm --force --stop