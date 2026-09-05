// Copyright 2024 the libevm authors.
//
// The libevm additions to go-ethereum are free software: you can redistribute
// them and/or modify them under the terms of the GNU Lesser General Public License
// as published by the Free Software Foundation, either version 3 of the License,
// or (at your option) any later version.
//
// The libevm additions are distributed in the hope that they will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU Lesser
// General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see
// <http://www.gnu.org/licenses/>.

// The precompilegen binary generates code for creating EVM precompiles
// conforming to arbitrary Solidity interfaces.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"io"
	"os"
	"sort"
	"strings"
	"text/template"

	"github.com/ava-labs/libevm/accounts/abi"

	_ "embed"
)

type config struct {
	in, out string
	pkg     string
}

func main() {
	var c config
	flag.StringVar(&c.in, "in", "", `Input ABI file or empty for stdin.`)
	flag.StringVar(&c.out, "out", "", `Output Go file or empty for stdout.`)
	flag.StringVar(&c.pkg, "package", "", "Generated package name.")
	flag.Parse()

	in := os.Stdin
	if c.in != "" {
		in = mustOpen(c.in, os.O_RDONLY, 0)
	}
	out := os.Stdout
	if c.out != "" {
		out = mustOpen(c.out, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	}

	if err := c.generate(in, out); err != nil {
		exit(err)
	}
	fmt.Fprintln(os.Stderr, "Generated precompile")
}

func mustOpen(name string, flag int, perm os.FileMode) *os.File {
	f, err := os.OpenFile(name, flag, perm) //nolint:gosec // User-provided file path is necessary behaviour, isolated to development environment.
	if err != nil {
		exit(err)
	}
	return f
}

func exit(a ...any) {
	fmt.Fprintln(os.Stderr, a...)
	os.Exit(1)
}

var (
	//go:embed precompile.go.tmpl
	rawTemplate string
	funcs       = template.FuncMap{
		"methods":     methods,
		"signature":   signature,
		"hex":         hex,
		"type":        goType,
		"args":        args,
		"interfaceID": interfaceID,
	}
	tmpl = template.Must(template.New("contract").Funcs(funcs).Parse(rawTemplate))
)

func (c *config) generate(abiJSON io.Reader, out io.WriteCloser) error {
	var jsonCopy bytes.Buffer
	parsed, err := abi.JSON(io.TeeReader(abiJSON, &jsonCopy))
	if err != nil {
		return fmt.Errorf("parse ABI: %v", err)
	}

	var buf bytes.Buffer
	data := struct {
		ABI     abi.ABI
		JSON    string
		Package string
	}{parsed, jsonCopy.String(), c.pkg}
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("execute template: %v", err)
	}

	src, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("format source: %v", err)
	}
	if _, err := out.Write(src); err != nil {
		return fmt.Errorf("write output: %v", err)
	}
	return out.Close()
}

func methods(a abi.ABI) []abi.Method {
	methods := make([]abi.Method, 0, len(a.Methods))
	for _, m := range a.Methods {
		methods = append(methods, m)
	}
	sort.Slice(methods, func(i, j int) bool {
		return methods[i].Name < methods[j].Name
	})
	return methods
}

func signature(m abi.Method) string {
	in := append([]string{"vm.PrecompileEnvironment"}, asGoArgs(m.Inputs)...)
	out := append(asGoArgs(m.Outputs), "error")
	return fmt.Sprintf("%s(%s) (%s)", m.Name, strings.Join(in, ","), strings.Join(out, ","))
}

// interfaceID returns the EIP-165 interface ID of the methods.
func interfaceID(a abi.ABI) abi.Selector {
	var id abi.Selector
	for _, m := range a.Methods {
		id ^= m.Selector()
	}
	return id
}

func asGoArgs(args abi.Arguments) []string {
	goArgs := make([]string, len(args))
	for i, a := range args {
		goArgs[i] = a.Type.GetType().String()
	}
	return goArgs
}

func hex(x any) string {
	return fmt.Sprintf("%#x", x)
}

func goType(a abi.Argument) string {
	return a.Type.GetType().String()
}

func args(prefix string, n int, withEnv, withErr bool) string {
	a := make([]string, n)
	for i := 0; i < n; i++ {
		a[i] = fmt.Sprintf("%s%d", prefix, i)
	}
	if withEnv {
		a = append([]string{"env"}, a...)
	}
	if withErr {
		a = append(a, "err")
	}
	return strings.Join(a, ", ")
}
