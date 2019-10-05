package framework

import (
	"flag"
	"fmt"
	"log"
	"reflect"
	"strings"
	"testing"
	"time"

	//"github.com/docker/docker/api/types"

	"github.com/smallinsky/mtf/framework/context"
	"github.com/smallinsky/mtf/pkg/cert"
)

func NewSuite(m *testing.M) *Suite {
	flag.Parse()
	return newSuite(m.Run)
}

type Suite struct {
	mRunFn   runFn
	settings Settings
}

func (s *Suite) Run() {

	kvpair, err := cert.GenCert([]string{"localhost", "host.docker.internal"})
	if err != nil {
		log.Fatalf("[ERR] failed to generate certs")
	}
	cert.WriteCert(kvpair)
	components := s.getComponents()

	start := time.Now()
	fmt.Println("=== PREPERING TEST ENV")
	components.Start()
	start = time.Now()
	fmt.Printf("=== PREPERING TEST ENV DONE - %v\n\n", time.Now().Sub(start))
	s.mRunFn()
	fmt.Printf("=== TEST RUN DONE - %v\n", time.Now().Sub(start))
	components.Stop()
}

type runFn func() int

func newSuite(run runFn) *Suite {
	return &Suite{
		mRunFn: run,
	}
}

func Run(t *testing.T, i interface{}) {
	if v, ok := i.(interface{ Init(*testing.T) }); ok {
		v.Init(t)
	}
	context.CreateDirectory()

	for _, test := range getInternalTests(i) {
		t.Run(test.Name, test.F)
	}
}

func getInternalTests(i interface{}) []testing.InternalTest {
	var tests []testing.InternalTest
	v := reflect.ValueOf(i)
	if v.Type().Kind() != reflect.Ptr && v.Type().Kind() != reflect.Struct {
		panic("arg is not a ptr to a struct")
	}
	for i := 0; i < v.Type().NumMethod(); i++ {
		tm := v.Type().Method(i)
		if !strings.HasPrefix(tm.Name, "Test") {
			continue
		}
		m := v.Method(i)
		if _, ok := m.Interface().(func(*testing.T)); !ok {
			continue
		}
		tests = append(tests, testing.InternalTest{
			Name: tm.Name,
			F: func(t *testing.T) {
				// create test dir
				context.CreateTestContext(t)
				m.Call([]reflect.Value{reflect.ValueOf(t)})
				context.RemoveTextContext(t)
				// get all port and run cleanup func
			},
		})
	}
	return tests
}
