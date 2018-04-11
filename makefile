build: 
	go build main.go 
run: 
	./main 
prof: 
	./main -cpuprofile=cpu.prof 
memprof: 
	./main -memprofile=mem.prof 
benchmark: 
	./main -t=100 -B=1000 -lambda=10 -keysize=2048 -cpuprofile=t_100_B_1000_lambda_10_key_2048_cpu.prof 

