.PHONY: all clean run

all: scanner

scanner: scanner.go
	go build -o scanner scanner.go

run: scanner
	./scanner input.pas

clean:
	rm -f scanner
