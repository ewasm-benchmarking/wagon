// Copyright 2017 The go-interpreter Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/go-interpreter/wagon/exec"
	"github.com/go-interpreter/wagon/validate"
	"github.com/go-interpreter/wagon/wasm"
)

func main() {
	log.SetPrefix("wasm-run: ")
	log.SetFlags(0)

	verbose := flag.Bool("v", false, "enable/disable verbose mode")
	verify := flag.Bool("verify-module", false, "run module verification")
	wasmFuncName := flag.String("func-name", "main", "called function name")

	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	wasm.SetDebugMode(*verbose)

	run(os.Stdout, flag.Arg(0), *wasmFuncName, *verify)
}

func run(w io.Writer, fname string, wasmFuncName string, verify bool) {
	f, err := os.Open(fname)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	readStart := time.Now()

	m, err := wasm.ReadModule(f, importer)
	if err != nil {
		log.Fatalf("could not read module: %v", err)
	}

	if verify {
		err = validate.VerifyModule(m)
		if err != nil {
			log.Fatalf("could not verify module: %v", err)
		}
	}

	if m.Export == nil {
		log.Fatalf("module has no export section")
	}

	vm, err := exec.NewVM(m)
	if err != nil {
		log.Fatalf("could not create VM: %v", err)
	}
	vm.RecoverPanic = true

	wasmFuncId, res := m.Export.Entries[wasmFuncName]
	if !res {
		log.Fatalf("could not find export function")
	}

	readElapsed := time.Since(readStart)
	fmt.Printf("parse time: %s\n", readElapsed)
	execStart := time.Now()

	o, err := vm.ExecCode(int64(wasmFuncId.Index))
	if err != nil {
		log.Fatalf("could not execute requested function: %v", err)
	}

	fmt.Fprintf(w, "%[1]v (%[1]T)\n", o)

	execElapsed := time.Since(execStart)
	fmt.Printf("exec time: %s\n", execElapsed)
}

func importer(name string) (*wasm.Module, error) {
	f, err := os.Open(name + ".wasm")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	m, err := wasm.ReadModule(f, nil)
	if err != nil {
		return nil, err
	}
	err = validate.VerifyModule(m)
	if err != nil {
		return nil, err
	}
	return m, nil
}
