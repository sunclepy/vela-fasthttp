package fasthttp

import (
	"fmt"
	"github.com/vela-security/vela-public/lua"
)

func (hd *handle) String() string                         { return fmt.Sprintf("fasthttp.handle %p", hd) }
func (hd *handle) Type() lua.LValueType                   { return lua.LTObject }
func (hd *handle) AssertFloat64() (float64, bool)         { return 0, false }
func (hd *handle) AssertString() (string, bool)           { return "", false }
func (hd *handle) AssertFunction() (*lua.LFunction, bool) { return nil, false }
func (hd *handle) Peek() lua.LValue                       { return hd }

func (hd *handle) NewIndex(L *lua.LState, key string, val lua.LValue) {
	switch key {
	case "method":
		hd.method = val.String()

	case "code":
		hd.code = lua.IsInt(val)

	case "filter":
		hd.filter = toFilter(L, val)

	case "header":
		hd.header = toHeader(L, val)

	case "close":
		hd.close = lua.IsFunc(val)

	case "eof":
		hd.eof = lua.IsTrue(val)

	case "body":
		switch val.Type() {
		case lua.LTString:
			hd.body = compileHandleBody(val.String())

		case lua.LTFunction:
			cp := xEnv.P(val.(*lua.LFunction))
			cp.NRet = 0

			hd.body = func(ctx *RequestCtx) error {
				co := newLuaThread(ctx)
				return co.CallByParam(cp)
			}

		default:
			hd.body = func(ctx *RequestCtx) error {
				ctx.SetBodyString(val.String())
				return nil
			}
		}

	}
}

func newLuaHandle(L *lua.LState) int {
	val := L.Get(1)

	hd := newHandle("")

	switch val.Type() {
	case lua.LTNil:
		hd.code = 404
		hd.body = func(ctx *RequestCtx) error {
			ctx.SetBodyString("nobody")
			return nil
		}

	case lua.LTString:
		hd.code = 200
		hd.eof = true
		hd.body = compileHandleBody(val.String())

	case lua.LTTable:
		val.(*lua.LTable).Range(func(key string, val lua.LValue) {
			hd.NewIndex(L, key, val)
		})

	default:
		hd.code = 200
		hd.eof = true
		hd.body = compileHandleBody(val.String())

	}

	L.Push(hd)
	return 1
}
