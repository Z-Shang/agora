package stdlib

import (
	"bufio"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/PuerkitoBio/agora/runtime"
)

type OsMod struct {
	ctx *runtime.Ctx
	ob  runtime.Object
}

type file struct {
	runtime.Object
	f *os.File
	s *bufio.Scanner
}

func (o *OsMod) newFile(f *os.File) *file {
	ob := runtime.NewObject()
	of := &file{
		ob,
		f,
		nil,
	}
	ob.Set(runtime.String("Name"), runtime.String(f.Name()))
	ob.Set(runtime.String("Close"), runtime.NewNativeFunc(o.ctx, "os.File.Close", of.closeFile))
	ob.Set(runtime.String("ReadLine"), runtime.NewNativeFunc(o.ctx, "os.File.ReadLine", of.readLine))
	ob.Set(runtime.String("Seek"), runtime.NewNativeFunc(o.ctx, "os.File.Seek", of.seek))
	ob.Set(runtime.String("Write"), runtime.NewNativeFunc(o.ctx, "os.File.Write", of.write))
	ob.Set(runtime.String("WriteLine"), runtime.NewNativeFunc(o.ctx, "os.File.WriteLine", of.writeLine))
	return of
}

func (of *file) closeFile(args ...runtime.Val) runtime.Val {
	e := of.f.Close()
	if e != nil {
		panic(e)
	}
	return runtime.Nil
}

func (of *file) readLine(args ...runtime.Val) runtime.Val {
	if of.s == nil {
		of.s = bufio.NewScanner(of.f)
	}
	if of.s.Scan() {
		return runtime.String(of.s.Text())
	}
	if e := of.s.Err(); e != nil {
		panic(e)
	}
	return runtime.Nil
}

func (of *file) seek(args ...runtime.Val) runtime.Val {
	off := int64(0)
	if len(args) > 0 {
		off = args[0].Int()
	}
	rel := 0
	if len(args) > 1 {
		rel = int(args[1].Int())
	}
	n, e := of.f.Seek(off, rel)
	if e != nil {
		panic(e)
	}
	return runtime.Number(n)
}

func (of *file) write(args ...runtime.Val) runtime.Val {
	n := 0
	for _, v := range args {
		if v != runtime.Nil {
			m, e := of.f.WriteString(v.String())
			if e != nil {
				panic(e)
			}
			n += m
		}
	}
	return runtime.Number(n)
}

func (of *file) writeLine(args ...runtime.Val) runtime.Val {
	n := of.write(args...)
	m, e := of.f.WriteString("\n")
	if e != nil {
		panic(e)
	}
	return runtime.Number(int(n.Int()) + m)
}

func (o *OsMod) ID() string {
	return "os"
}

func (o *OsMod) Run(_ ...runtime.Val) (v runtime.Val, err error) {
	defer runtime.PanicToError(&err)
	if o.ob == nil {
		// Prepare the object
		o.ob = runtime.NewObject()
		o.ob.Set(runtime.String("PathSeparator"), runtime.String(os.PathSeparator))
		o.ob.Set(runtime.String("PathListSeparator"), runtime.String(os.PathListSeparator))
		o.ob.Set(runtime.String("DevNull"), runtime.String(os.DevNull))
		o.ob.Set(runtime.String("Exec"), runtime.NewNativeFunc(o.ctx, "os.Exec", o.os_Exec))
		o.ob.Set(runtime.String("Exit"), runtime.NewNativeFunc(o.ctx, "os.Exit", o.os_Exit))
		o.ob.Set(runtime.String("Getenv"), runtime.NewNativeFunc(o.ctx, "os.Getenv", o.os_Getenv))
		o.ob.Set(runtime.String("Getwd"), runtime.NewNativeFunc(o.ctx, "os.Getwd", o.os_Getwd))
		o.ob.Set(runtime.String("ReadFile"), runtime.NewNativeFunc(o.ctx, "os.ReadFile", o.os_ReadFile))
		o.ob.Set(runtime.String("WriteFile"), runtime.NewNativeFunc(o.ctx, "os.WriteFile", o.os_WriteFile))
		o.ob.Set(runtime.String("Open"), runtime.NewNativeFunc(o.ctx, "os.Open", o.os_Open))
	}
	return o.ob, nil
}

func (o *OsMod) SetCtx(ctx *runtime.Ctx) {
	o.ctx = ctx
}

func (o *OsMod) os_Exit(args ...runtime.Val) runtime.Val {
	if len(args) == 0 {
		os.Exit(0)
	}
	os.Exit(int(args[0].Int()))
	return runtime.Nil
}

func (o *OsMod) os_Getenv(args ...runtime.Val) runtime.Val {
	runtime.ExpectAtLeastNArgs(1, args)
	return runtime.String(os.Getenv(args[0].String()))
}

func (o *OsMod) os_Getwd(args ...runtime.Val) runtime.Val {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return runtime.String(pwd)
}

func (o *OsMod) os_Exec(args ...runtime.Val) runtime.Val {
	runtime.ExpectAtLeastNArgs(1, args)
	c := exec.Command(args[0].String(), toString(args[1:])...)
	b, e := c.CombinedOutput()
	if e != nil {
		panic(e)
	}
	return runtime.String(b)
}

func (o *OsMod) os_ReadFile(args ...runtime.Val) runtime.Val {
	runtime.ExpectAtLeastNArgs(1, args)
	b, e := ioutil.ReadFile(args[0].String())
	if e != nil {
		panic(e)
	}
	return runtime.String(b)
}

func (o *OsMod) os_WriteFile(args ...runtime.Val) runtime.Val {
	runtime.ExpectAtLeastNArgs(1, args)
	f, e := os.Create(args[0].String())
	if e != nil {
		panic(e)
	}
	defer f.Close()
	n := 0
	for _, v := range args[1:] {
		if v != runtime.Nil {
			m, e := f.WriteString(v.String())
			if e != nil {
				panic(e)
			}
			n += m
		}
	}
	return runtime.Number(n)
}

func (o *OsMod) os_Open(args ...runtime.Val) runtime.Val {
	runtime.ExpectAtLeastNArgs(1, args)
	nm := args[0].String()
	flg := "r" // defaults to read-only
	if len(args) > 1 {
		// Second arg is the flag (r - w - a)
		flg = args[1].String()
	}
	flgi := os.O_RDONLY
	switch flg {
	case "w":
		flgi = os.O_WRONLY
	case "rw":
		flgi = os.O_RDWR
	case "a":
		flgi = os.O_APPEND
	}
	f, e := os.OpenFile(nm, flgi, 0)
	if e != nil {
		panic(e)
	}
	return o.newFile(f)
}

func toString(args []runtime.Val) []string {
	s := make([]string, len(args))
	for i, a := range args {
		s[i] = a.String()
	}
	return s
}