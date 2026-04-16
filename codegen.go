package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const Reg = 10

type Item struct {
	Mode int
	Type *TypeDesc
	A int64
	name string
}

var pc int
var useWrite bool
var lastAssigned string
var codeLines []string

func initCodegen() {
	pc = 0
	useWrite = false
	lastAssigned = ""
	codeLines = nil
}

func header() {}

func emit(s string) {
	codeLines = append(codeLines, s)

	pc++
}

func load(x *Item) {
	if x.Mode == Reg {
		return
	}

	switch x.Mode {
		case Const:
			emit("    movq $" + strconv.FormatInt(x.A, 10) + ", %rax")
		case Var:
			emit("    movq " + x.name + "(%rip), %rax")
	}

	x.Mode = Reg
}

func makeConstItem(x *Item, typ *TypeDesc, val int64) {
	x.Mode = Const
	x.Type = typ
	x.A = val
}

func makeItem(x *Item, obj *ObjDesc) {
	x.Mode = obj.Class
	x.Type = obj.Type
	x.A = obj.Val
	x.name = obj.Name
}

func neg(x *Item) {
	if x.Mode == Const {
		x.A = -x.A

		return
	}

	load(x)

	emit("    negq %rax")
}

func addOp(op int, x, y *Item) {
	if x.Mode == Const && y.Mode == Const {
		if op == symPlus {
			x.A += y.A
		} else {
			x.A -= y.A
		}

		return
	}

	load(x)

	var rhs string

	switch y.Mode {
		case Const:
			rhs = "$" + strconv.FormatInt(y.A, 10)
		case Var:
			rhs = y.name + "(%rip)"
		default:
			return
	}

	if op == symPlus {
		emit("    addq " + rhs + ", %rax")
	} else {
		emit("    subq " + rhs + ", %rax")
	}

	x.Mode = Reg
}

func store(x, y *Item) {
	load(y)

	emit("    movq %rax, " + x.name + "(%rip)")

	lastAssigned = x.name
}

func writeCall(x *Item) {
	emit("    movq " + x.name + "(%rip), %rdi")
	emit("    call print_int")

	useWrite = true
}

func checkRegs() {}

func close(base string) {
	asm := buildAssembly()

	asmFile := base + ".s"
	objFile := base + ".o"
	exeFile := base

	if err := os.WriteFile(asmFile, []byte(asm), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "cannot write %s: %v\n", asmFile, err)

		os.Exit(1)
	}

	fmt.Println("→", asmFile)

	runCmd("as", "--64", "-o", objFile, asmFile)

	fmt.Println("→", objFile)

	runCmd("ld", "-o", exeFile, objFile)

	fmt.Println("→", exeFile)
}

func buildAssembly() string {
	var sb strings.Builder

	obj := progScope.Next

	for obj != nil {
		if obj.Class == Var {
			sb.WriteString("    .section .bss\n")
			break
		}

		obj = obj.Next
	}

	for obj := progScope.Next; obj != nil; obj = obj.Next {
		if obj.Class == Var {
			sb.WriteString("    .align 8\n")
			sb.WriteString(obj.Name + ":\n    .zero 8\n")
		}
	}

	sb.WriteString("\n    .section .text\n    .globl _start\n\n")

	if useWrite {
		sb.WriteString(printIntASM + "\n\n")
	}

	sb.WriteString("_start:\n")

	for _, l := range codeLines {
		sb.WriteString(l + "\n")
	}

	sb.WriteString("\n    # sys_exit(60)\n")

	if lastAssigned != "" {
		sb.WriteString("    movq " + lastAssigned + "(%rip), %rdi\n")
	} else {
		sb.WriteString("    movq $0, %rdi\n")
	}

	sb.WriteString("    movq $60, %rax\n    syscall\n")

	return sb.String()
}

func runCmd(name string, args ...string) {
	cmd := exec.Command(name, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s failed: %v\n", name, err)

		os.Exit(1)
	}
}

const printIntASM = `print_int:
    pushq   %rbp
    movq    %rsp, %rbp
    subq    $32, %rsp
    movq    %rdi, %r8
    movq    %rdi, %rax
    testq   %rax, %rax
    jge     .Lpi_pos
    negq    %rax
.Lpi_pos:
    leaq    -1(%rbp), %rsi
    movb    $10, (%rsi)
    decq    %rsi
    movq    $10, %rcx
.Lpi_loop:
    xorq    %rdx, %rdx
    divq    %rcx
    addb    $48, %dl
    movb    %dl, (%rsi)
    decq    %rsi
    testq   %rax, %rax
    jnz     .Lpi_loop
    incq    %rsi
    testq   %r8, %r8
    jge     .Lpi_write
    decq    %rsi
    movb    $45, (%rsi)
.Lpi_write:
    leaq    -1(%rbp), %rdx
    subq    %rsi, %rdx
    incq    %rdx
    movq    $1,  %rax
    movq    $1,  %rdi
    syscall
    addq    $32, %rsp
    popq    %rbp
    ret`

// https://people.inf.ethz.ch/wirth/ProjectOberon/Sources/ORG.Mod.txt