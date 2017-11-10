# try running this script to start three servers
# make an example query thrice to see the speed up
# 
# time curl http://127.0.0.1:8080/define/friend
#
# and again, and you would notice that in the real time is reduced due to caching
#
# time curl http://127.0.0.1:8080/define/friend
# 
# try with another port
#
# time curl http://127.0.0.1:8081/define/friend





go run example4.go -addr=:8080 -pool=http://127.0.0.1:8080,http://127.0.0.1:8081,http://127.0.0.1:8082 &
go run example4.go -addr=:8081 -pool=http://127.0.0.1:8081,http://127.0.0.1:8080,http://127.0.0.1:8082 &
go run example4.go -addr=:8082 -pool=http://127.0.0.1:8082,http://127.0.0.1:8080,http://127.0.0.1:8081 &
