build: 
	go build main.go 
run: 
	./main 
prof: 
	./main -cpuprofile=cpu.prof 
memprof: 
	./main -memprofile=mem.prof 
benchmark: 
	./main -t=1000 -B=10000 -lambda=100 -keysize=2048 -cpuprofile=t_1000_B_10000_lambda_100_key_2048_cpu.prof 
full: 
	go run setup/setup.go -t=1000 -B=10000 -lambda=100 -keysize=512
	go run evalinit/evalinit.go -t=1000 -B=10000 -lambda=100 -keysize=512
	go run veriinit/veriinit.go -t=1000 -B=10000 -lambda=100 -keysize=512
	go run eval/eval.go -t=1000 -B=10000 -lambda=100 -keysize=512
	go run verify/verify.go -t=1000 -B=10000 -lambda=100 -keysize=512


