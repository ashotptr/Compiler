COMPILER = ./compiler
SOURCES  = scanner.go symtab.go codegen.go parser.go main.go

all:
	go build -o compiler $(SOURCES)

test: all test.pas
	$(COMPILER) test.pas
	./test; echo $$?

clean:
	rm -f compiler test test.s test.o

.PHONY: all test clean