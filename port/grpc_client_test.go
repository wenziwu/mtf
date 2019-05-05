package port

import (
	"context"
	"reflect"
	"sync"
	"testing"

	"google.golang.org/grpc"
)

func TestGrpcClientPort(t *testing.T) {
	port := ClientPort{
		emd: map[reflect.Type]EndpointRespTypePair{
			reflect.TypeOf((*FirstRequest)(nil)): EndpointRespTypePair{
				Endpoint: "FirstMessageHandler",
				RespType: reflect.TypeOf((*FirstResponse)(nil)),
			},
		},
		callResultC: make(chan callResult),
		conn:        &mockConnection{t: t},
	}

	t.Run("SendReceiveOneMessgeSameType", func(t *testing.T) {
		port.Send(&FirstRequest{
			ID: 1,
		})
		port.Receive(&FirstResponse{
			ID: 1,
		})
	})

	t.Run("SendReceiveTwoMessageSameType", func(t *testing.T) {
		t.Skipf("fix async call and queue messages")
		port.Send(&FirstRequest{
			ID: 1,
		})
		port.Send(&FirstRequest{
			ID: 2,
		})

		//TODO: Fix async order based on send call
		port.Receive(&FirstResponse{
			ID: 1,
		})
		port.Receive(&FirstResponse{
			ID: 2,
		})
	})

	t.Run("SendReciveParaller", func(t *testing.T) {
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				i := i

				port.Send(&FirstRequest{
					ID: i,
				})
				port.Receive(&FirstResponse{
					ID: i,
				})
			}()
			wg.Wait()
		}
	})

	t.Run("RecieveMatchFn", func(t *testing.T) {
		port.Send(&FirstRequest{
			ID: 10,
		})
		port.ReceiveMatch(
			func(r *FirstResponse) {
				if r.ID != 10 {
					t.Fatalf("expected response id = 10 but got: %v", r.ID)
				}
			},
		)
	})
}

type mockConnection struct {
	t   *testing.T
	err error
}

func (m *mockConnection) Invoke(ctx context.Context, method string, args, reply interface{}, opts ...grpc.CallOption) error {
	switch t := args.(type) {
	case *FirstRequest:
		resp := FirstResponse{ID: t.ID}
		reflect.ValueOf(reply).Elem().Set(reflect.ValueOf(resp))
	case *SecondRequest:
		resp := FirstResponse{ID: t.ID}
		reflect.ValueOf(reply).Elem().Set(reflect.ValueOf(resp))
	}

	return m.err
}
func (m *mockConnection) Close() error {
	return nil
}

type FirstRequest struct {
	ID int
}

type FirstResponse struct {
	ID int
}

type SecondRequest struct {
	ID int
}

type SecondResponse struct {
	ID int
}