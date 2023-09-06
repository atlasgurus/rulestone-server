build:
	go build -o ./build/rulestone ./main/main.go

clean:
	@echo "Cleaning up"
	rm -rf ./build

.PHONY: build
