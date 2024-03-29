.PHONY: default build make windows_build_386 linux_build clean

default:
	@echo 'Usage of make: [ build | clean | windows_build | linux_build ]'

build:
	@go build -o ./build/main .

linux_build:
	@rm -f main
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./build/main .

windows_build:
	@rm -f main.exe
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o ./build/main.exe .

windows_build_386:
	@rm -f main.exe
	@CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -o ./build/main.exe .

clean: 
	@rm -f ./build/main
	@rm -f ./build/main.exe
	@rm -f ./build/logs/*.log