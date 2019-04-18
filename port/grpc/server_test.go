package grpc

import (
	"context"
	"fmt"
	"testing"
	"time"

	"google.golang.org/grpc"

	"github.com/smallinsky/mtf/e2e/proto/oracle"
)

func TestGRPCServer(t *testing.T) {
	svr := NewServer((*oracle.OracleServer)(nil), ":9999")
	conn, err := grpc.Dial("localhost:9999", grpc.WithInsecure())
	if err != nil {
		t.Fatal("fialed to dial echo addres: ", err)
	}
	defer conn.Close()
	client := oracle.NewOracleClient(conn)

	t.Run("SingeCall", func(t *testing.T) {
		go func() {
			svr.Receive(&oracle.AskDeepThroughRequest{
				Data: "Ultimate question",
			})

			svr.Send(&oracle.AskDeepThroughRespnse{
				Data: "42",
			})
		}()

		resp, err := client.AskDeepThrough(context.Background(), &oracle.AskDeepThroughRequest{
			Data: "Ultimate question",
		})
		if err != nil {
			t.Fatal("faield to ask deep through: ", err)
		}

		if got, exp := resp.GetData(), "42"; got != exp {
			t.Fatalf("Got: '%v' Expected: '%v'", got, exp)
		}
	})

	t.Run("MultileSeqCalls", func(t *testing.T) {
		const (
			N = 100
		)

		go func() {
			for i := 0; i < N; i++ {
				svr.Receive(&oracle.AskDeepThroughRequest{
					Data: fmt.Sprintf("Request: %v", i),
				})
				svr.Send(&oracle.AskDeepThroughRespnse{
					Data: fmt.Sprintf("Response: %v", i),
				})
			}
		}()

		for i := 0; i < N; i++ {
			resp, err := client.AskDeepThrough(context.Background(), &oracle.AskDeepThroughRequest{
				Data: fmt.Sprintf("Request: %v", i),
			})
			if err != nil {
				t.Fatal("failed to ask deep through: ", err)
			}

			if got, exp := resp.GetData(), fmt.Sprintf("Response: %v", i); got != exp {
				t.Fatalf("Got: '%v' Expected: '%v'", got, exp)
			}
		}
	})
}

func TestGRPCServerStart(t *testing.T) {
	t.Run("Delay", func(t *testing.T) {
		conn, err := grpc.Dial("localhost:9991", grpc.WithInsecure())
		if err != nil {
			t.Fatal("fialed to dial echo addres: ", err)
		}

		done := make(chan struct{})
		defer close(done)
		go func() {
			prevousState := ""
			for {
				select {
				case <-done:
					return
				default:
					currState := conn.GetState().String()
					if currState != prevousState {
						t.Logf("%v Diff curr: '%v'", time.Now(), currState)
						prevousState = currState
						continue
					}
					time.Sleep(time.Microsecond * 100)
				}
			}
		}()
		defer conn.Close()
		client := oracle.NewOracleClient(conn)

		// Value 1s are causing causes client grpc.Dial error call.
		time.Sleep(time.Second * 5)
		svr := NewServer((*oracle.OracleServer)(nil), ":9991")
		go func() {
			svr.Receive(&oracle.AskDeepThroughRequest{
				Data: "Ultimate question",
			})

			svr.Send(&oracle.AskDeepThroughRespnse{
				Data: "42",
			})
		}()

		time.Sleep(time.Second)
		resp, err := client.AskDeepThrough(context.Background(), &oracle.AskDeepThroughRequest{
			Data: "Ultimate question",
		})
		if err != nil {
			t.Fatal("faield to ask deep through: ", err)
		}

		if got, exp := resp.GetData(), "42"; got != exp {
			t.Fatalf("Got: '%v' Expected: '%v'", got, exp)
		}

	})
}