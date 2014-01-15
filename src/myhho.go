package main

import (
	"go/parser"
	"go/ast"
	"go/token"
	"fmt"
	"reflect"
	"strings"
)

const hellogo string = `
package main

func main() {
	b := 5
	c := 4
	print("Testing", b, ">=", c)
	if (b >= c) {
		println("yep")
	} else {
		println("nope")
	}
}
`

func getOpFromKind(t token.Token) string {
	s := ""
	switch t {
	case token.STRING:
		s = "String"
	case token.CHAR:
		s = "String"
	case token.INT:
		s = "Int"
	case token.FLOAT:
		s = "Double"
	case token.ADD:
		s = "Add"
	case token.SUB:
		s = "Sub"
	case token.MUL:
		s = "Mul"
	case token.QUO:
		s = "Div"
	case token.REM:
		s = "Mod"
	case token.AND:     // &
		s = "BitAnd"
	case token.OR:        // |
		s = "BitOr"
	case token.XOR:     // ^
		s = "BitXor"
	case token.SHL:     // <<
		s = "Shl"
	case token.SHR:     // >>
		s = "Shr"
	case token.EQL:    // ==
		s = "Eq"
	case token.LSS:    // <
		s = "Lt"
	case token.GTR:    // >
		s = "Gt"
	case token.NOT:   // !
		s = "Not"
	case token.NEQ:      // !=
		s = "Neq"
	case token.LEQ:      // <=
		s = "Lte"
	case token.GEQ:      // >=
		s = "Gte"
	default:
		panic("Token not supported")

	}
	return s
}

type Assembler struct {
	hhas            string
	cur_label,
	indent          int
	skip_next_ident,
	in_assign,
	in_lhs            bool
}

func NewAssembler() *Assembler {
	a := new(Assembler)
	a.hhas = ""
	a.indent = 0
	a.cur_label = 0
	a.skip_next_ident = false
	return a
}

func (a *Assembler) emit(fstring string, args ...interface{}) (string) {
	ind := strings.Repeat("    ", a.indent)
	str := fmt.Sprintf(fstring, args...)
	return fmt.Sprintf("%s%s\n", ind, str)
}

func (a *Assembler) Print() {
	fmt.Println(a.hhas)
}

func buildArgList(n *ast.FuncDecl) string {
	s := ""
	args := make([]string, 0)
	for _, x := range n.Type.Params.List {
		for _, y := range x.Names {
			args = append(args, y.Name)
		}
	}
	s = strings.Join(args, ", ")
	return s
}

func (a *Assembler) EmitAssignStmt(n *ast.AssignStmt) {
	a.in_assign = true
	a.ParseNode(n.Rhs[0])
	a.in_lhs = true
	a.ParseNode(n.Lhs[0])
	a.in_lhs = false
	a.in_assign = false
}

func (a *Assembler) EmitBasicLit(n *ast.BasicLit) {
	a.hhas += a.emit(getOpFromKind(n.Kind)+" %s", n.Value)
}

func (a *Assembler) EmitBinaryExpr(n *ast.BinaryExpr) {
	a.ParseNode(n.Y)
	a.ParseNode(n.X)
	a.hhas += a.emit(getOpFromKind(n.Op))

}

func (a *Assembler) EmitBlockStmt(n *ast.BlockStmt) {
	for _, x := range n.List {
		a.ParseNode(x)
	}
}

func (a *Assembler) EmitCallArgs(n []ast.Expr) {
	for i, arg := range n {
		switch v := arg.(type) {
		case *ast.BasicLit:
			a.hhas += a.emit("FPassC %d %s", i, v.Value)
		case *ast.Ident:
			a.hhas += a.emit("FPassL %d %s", i, v.Name)
		default:
			fmt.Printf("Unrecognized type: %s\n", v)
		}
	}
	a.hhas += a.emit("FCall %d", len(n))
}

func (a *Assembler) EmitCallExpr(n *ast.CallExpr) {
	fname := n.Fun.(*ast.Ident).Name
	printargs := func(args []ast.Expr) {
		for _, arg := range args {
			ast.Inspect(arg, a.ParseNode)
			a.hhas += a.emit("Print")
			a.hhas += a.emit("PopC")
		}
	}
	if (fname == "print" || fname == "println") {
		printargs(n.Args)
		if (fname == "println") {
			a.hhas += a.emit("String \"\\n\"")
			a.hhas += a.emit("Print")
		}
	} else {
		a.hhas += a.emit("FPushFuncD %d \"%s\"", len(n.Args), fname)
		a.EmitCallArgs(n.Args)
	}
}

func (a *Assembler) EmitFile(n *ast.File) {
	//TODO: Bunch of stuff relating to packages, etc
	main := new(ast.CallExpr)
	main.Fun = new(ast.Ident)
	main.Fun.(*ast.Ident).Name = "main"
	a.hhas += a.emit(".main {")
	a.indent++
	a.EmitCallExpr(main)
	a.hhas += a.emit("PopR")
	a.hhas += a.emit("Int 0")
	a.hhas += a.emit("RetC")
	a.indent--
	a.hhas += a.emit("}\n")
	a.skip_next_ident = true
}

func (a *Assembler) EmitFuncBody(n *ast.BlockStmt) {
	for _, x := range n.List {
		a.ParseNode(x)
	}
}

func (a *Assembler) EmitFuncDecl(n *ast.FuncDecl) {
	args := buildArgList(n)
	a.hhas += a.emit(".function %s(%s) {", n.Name.Name, args)
	a.indent++
	a.EmitFuncBody(n.Body)
	a.hhas += a.emit("RetC")
	a.indent--
	a.hhas += a.emit("}")
}

func (a *Assembler) EmitIdent(n *ast.Ident) {
	if (a.in_assign) {
		if (a.in_lhs) {
			a.hhas += a.emit("SetL $%s", n.Name)
			a.hhas += a.emit("PopC")
			return
		}
	}
	a.hhas += a.emit("CGetL $%s", n.Name)
}

func (a *Assembler) EmitIfStmt(n *ast.IfStmt) {
	a.ParseNode(n.Cond)
	label := a.getNextLabel()
	elseLabel := label + "_else"
	endLabel := label + "_end"

	emitLabel := func(l string) { a.hhas += a.emit("%s:", l) }

	a.hhas += a.emit("JmpNZ %s", elseLabel)
	emitLabel(label)
	a.indent++
	a.ParseNode(n.Body)
	a.hhas += a.emit("Jmp %s", endLabel)
	a.indent--
	emitLabel(elseLabel)
	a.indent++
	a.ParseNode(n.Else)
	a.indent--
	emitLabel(endLabel)
}

func (a *Assembler) EmitParenExpr(n *ast.ParenExpr) {
	a.ParseNode(n.X)
}

func (a *Assembler) getNextLabel() (lbl string) {
	lbl = fmt.Sprintf("label_%d", a.cur_label)
	a.cur_label++
	return
}

func (a *Assembler) EmitReturnStmt(n *ast.ReturnStmt) {
	getRetVal := func(e ast.Expr) string {
		switch v := e.(type) {
		case *ast.BinaryExpr:
			x := v.X.(*ast.BasicLit)
			y := v.Y.(*ast.BasicLit)
			s := ""
			s += a.emit(getOpFromKind(x.Kind)+"%s", x.Value)
			s += a.emit(getOpFromKind(y.Kind)+"%s", y.Value)
			s += a.emit(getOpFromKind(v.Op))
			return s
		case *ast.BasicLit:
			return a.emit(getOpFromKind(v.Kind)+"%s", v.Value)
		case *ast.Ident:
			return a.emit("CGetL $%s", v.Name)
		default:
			panic(fmt.Sprintf("Unexpected return: %s\n", reflect.TypeOf(v)))
		}
	}

	if (len(n.Results) == 0) {
		a.hhas += a.emit("RetC")
		return
	}
	if (len(n.Results) == 1) {
		a.hhas += getRetVal(n.Results[0])
	}
	if (len(n.Results) > 1) {
		a.hhas += a.emit("NewArray")
		//		for _, r := range n.Results {

		//		}
	}
	a.hhas += a.emit("RetC")

}

func (a *Assembler) ParseNode(n ast.Node) bool {
	if (n == nil) {
		return false
	}
	switch v := n.(type) {
	case *ast.AssignStmt:
		a.EmitAssignStmt(v)
	case *ast.BinaryExpr:
		a.EmitBinaryExpr(v)
	case *ast.BasicLit:
		a.EmitBasicLit(v)
	case *ast.BlockStmt:
		a.EmitBlockStmt(v)
	case *ast.CallExpr:
		a.EmitCallExpr(v)
	case *ast.ExprStmt:
		a.ParseNode(v.X)
	case *ast.File:
		a.EmitFile(v)
	case *ast.FuncDecl:
		a.EmitFuncDecl(v)
		return false
	case *ast.Ident:
		if (!a.skip_next_ident) {
			a.EmitIdent(v)
		} else {
			a.skip_next_ident = false
		}
	case *ast.IfStmt:
		a.EmitIfStmt(v)
	case *ast.ParenExpr:
		a.EmitParenExpr(v)
	case *ast.ReturnStmt:
		a.EmitReturnStmt(v)
		return false
	default:
		// v is a ast.Node
		fmt.Println("Not implemented:", reflect.TypeOf(v))
	}
	return true
}

func main() {
	f := token.NewFileSet()
	t, err := parser.ParseFile(f, "hello.go", hellogo, 0)
	if (err != nil) {
		print(err)
	}

	a := NewAssembler()
	ast.Inspect(t, a.ParseNode)
	//ast.Print(f, t)
	a.Print()
}
