curl --data "password=angryMonkey" http://localhost:8080/hash
echo ''
curl -X POST localhost:8080/hash/ -F "password"="123"
echo ''
curl -X POST localhost:8080/hash -F "password"="something" 
echo ''

sleep 5

curl localhost:8080/hash/1 
echo ''
curl localhost:8080/hash/2 
echo ''
curl localhost:8080/hash/3 
echo ''
curl -i localhost:8080/hash/4
echo ''

curl localhost:8080/stats
echo ''

curl localhost:8080/shutdown
echo''
