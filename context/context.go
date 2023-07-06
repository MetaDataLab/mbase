// Author: CJey Hou<cjey.hou@ucloud.cn>
package context

import (
	gcontext "context"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

// variable alias
var (
	Canceled         = gcontext.Canceled
	DeadlineExceeded = gcontext.DeadlineExceeded
)

// type alias
type (
	CancelFunc = gcontext.CancelFunc
)

// Context extend default context package, make it better
// It should has name, location, environment and logger
type Context interface {
	// composite
	gcontext.Context

	// Fork return a copied context, if there is a new goroutine generated
	// it will use my name with a sequential number suffix started from 1
	Fork() Context
	// At return a copied context, specify the current location where it is in,
	// it should chain all locations start from root
	At(location string) Context
	ForkAt(location string) Context
	// Reborn will use gcontext.Background() instead of internal context,
	// it used for escaping internal context's cancel request
	Reborn() Context
	// RebornWith will use specified context instead of internal context,
	// it used for escaping internal context's cancel request
	RebornWith(gcontext.Context) Context
	// Name return my logger's name
	Name() string
	// Location return my logger's location
	Location() string

	// integrated base official context action
	WithCancel() (Context, CancelFunc)
	WithDeadline(time.Time) (Context, CancelFunc)
	WithTimeout(time.Duration) (Context, CancelFunc)
	WithValue(key, value interface{}) Context

	// Env return my env
	// WARN: env value and official context value are two diffrent things
	Env() Env
	// shortcut methods of my env
	Set(key, value interface{})
	Get(key interface{}) (value interface{}, ok bool)
	GetString(key interface{}) string
	GetInt(key interface{}) int
	GetUint(key interface{}) uint
	GetFloat(key interface{}) float64
	GetBool(key interface{}) bool

	// Logger return my logger
	Logger() Logger
	// shortcut methods of my logger
	Debug(msg string, kvs ...interface{})
	Info(msg string, kvs ...interface{})
	Warn(msg string, kvs ...interface{})
	Error(msg string, kvs ...interface{})
	Panic(msg string, kvs ...interface{})
	Fatal(msg string, kvs ...interface{})
}

type context struct {
	gctx    gcontext.Context
	tracker *uint64

	env    Env
	logger Logger
}

var _ Context = (*context)(nil)

// Generator 定义了一个Context的生成函数，每次调用都应当返回一个新的Context
type Generator = func() Context

// Simple return a very simple context, without name, without location,
// and use S() as internal logger
func Simple() Context {
	return New(
		gcontext.Background(),
		NewEnv(),
		NewLogger("", "", _S, nil, nil),
	)
}

// Sessinal 返回一个简单的Context，命名部分使用自动生成的uuid。
func Sessional() Context {
	return New(
		gcontext.Background(),
		NewEnv(),
		NewLogger(uuid.New().String(), "", _S, nil, nil),
	)
}

// Named 返回一个简单的Context，Context的名称可以通过参数name实现自定义。
func Named(name string) Context {
	return New(
		gcontext.Background(),
		NewEnv(),
		NewLogger(name, "", _S, nil, nil),
	)
}

// ToSimple 将给定的标准库Context对象作为Context的内部基础对象。
func ToSimple(gctx gcontext.Context) Context {
	return New(
		gctx,
		NewEnv(),
		NewLogger("", "", _S, nil, nil),
	)
}

// ToSesional 将给定的标准库Context对象对位Context的内部基础对象，并自动生成uuid作为name。
func ToSessional(gctx gcontext.Context) Context {
	return New(
		gctx,
		NewEnv(),
		NewLogger(uuid.New().String(), "", _S, nil, nil),
	)
}

// ToNamed 使用给定的标准库Context和名称列表构建一个Context。
func ToNamed(gctx gcontext.Context, name string) Context {
	return New(
		gctx,
		NewEnv(),
		NewLogger(name, "", _S, nil, nil),
	)
}

// New use an official Context, an Env and a Logger to generate a new Context.
// It will use default value if not given.
func New(gctx gcontext.Context, env Env, logger Logger) Context {
	if gctx == nil {
		gctx = gcontext.Background()
	}
	if env == nil {
		env = NewEnv()
	}
	if logger == nil {
		logger = NewLogger("", "", nil, nil, nil)
	}

	var tracker uint64
	return &context{
		gctx:    gctx,
		tracker: &tracker,

		env:    env,
		logger: logger,
	}
}

func (ctx *context) fork(name, location string) *context {
	return &context{
		gctx:    ctx.gctx,
		tracker: ctx.tracker,

		env:    ctx.env.Fork(),
		logger: ctx.logger.Fork(name, location),
	}
}

func (ctx *context) Deadline() (deadline time.Time, ok bool) {
	return ctx.gctx.Deadline()
}

func (ctx *context) Done() <-chan struct{} {
	return ctx.gctx.Done()
}

func (ctx *context) Err() error {
	return ctx.gctx.Err()
}

func (ctx *context) Value(key interface{}) interface{} {
	return ctx.gctx.Value(key)
}

func (ctx *context) Fork() Context {
	return ctx.ForkAt("")
}

func (ctx *context) At(location string) Context {
	return ctx.fork("", location)
}

func (ctx *context) ForkAt(location string) Context {
	var seq = atomic.AddUint64(ctx.tracker, 1)
	var newctx = ctx.fork(strconv.FormatUint(seq, 10), location)
	var tracker uint64
	newctx.tracker = &tracker
	return newctx
}

func (ctx *context) Reborn() Context {
	return ctx.RebornWith(gcontext.Background())
}

func (ctx *context) RebornWith(gctx gcontext.Context) Context {
	if gctx == nil {
		gctx = gcontext.Background()
	}
	var newctx = ctx.fork("", "")
	newctx.gctx = gctx
	return newctx
}

func (ctx *context) Name() string {
	return ctx.logger.Name()
}

func (ctx *context) Location() string {
	return ctx.logger.Location()
}

func (ctx *context) WithCancel() (Context, CancelFunc) {
	var newctx = ctx.fork("", "")
	var newgctx, f = gcontext.WithCancel(newctx.gctx)
	newctx.gctx = newgctx
	return newctx, f
}

func (ctx *context) WithDeadline(d time.Time) (Context, CancelFunc) {
	var newctx = ctx.fork("", "")
	var newgctx, f = gcontext.WithDeadline(newctx.gctx, d)
	newctx.gctx = newgctx
	return newctx, f
}

func (ctx *context) WithTimeout(timeout time.Duration) (Context, CancelFunc) {
	var newctx = ctx.fork("", "")
	var newgctx, f = gcontext.WithTimeout(newctx.gctx, timeout)
	newctx.gctx = newgctx
	return newctx, f
}

func (ctx *context) WithValue(key, value interface{}) Context {
	var newctx = ctx.fork("", "")
	newctx.gctx = gcontext.WithValue(newctx.gctx, key, value)
	return newctx
}

func (ctx *context) Env() Env {
	return ctx.env
}

func (ctx *context) Logger() Logger {
	return ctx.logger
}

func (ctx *context) Set(key, value interface{}) {
	ctx.env.Set(key, value)
}

func (ctx *context) Get(key interface{}) (value interface{}, ok bool) {
	return ctx.env.Get(key)
}

func (ctx *context) GetString(key interface{}) string {
	return ctx.env.GetString(key)
}

func (ctx *context) GetInt(key interface{}) int {
	return ctx.env.GetInt(key)
}

func (ctx *context) GetUint(key interface{}) uint {
	return ctx.env.GetUint(key)
}

func (ctx *context) GetFloat(key interface{}) float64 {
	return ctx.env.GetFloat(key)
}

func (ctx *context) GetBool(key interface{}) bool {
	return ctx.env.GetBool(key)
}

func (ctx *context) Debug(msg string, kvs ...interface{}) {
	ctx.logger.Debug(msg, kvs...)
}

func (ctx *context) Info(msg string, kvs ...interface{}) {
	ctx.logger.Info(msg, kvs...)
}

func (ctx *context) Warn(msg string, kvs ...interface{}) {
	ctx.logger.Warn(msg, kvs...)
}

func (ctx *context) Error(msg string, kvs ...interface{}) {
	ctx.logger.Error(msg, kvs...)
}

func (ctx *context) Panic(msg string, kvs ...interface{}) {
	ctx.logger.Panic(msg, kvs...)
}

func (ctx *context) Fatal(msg string, kvs ...interface{}) {
	ctx.logger.Fatal(msg, kvs...)
}
