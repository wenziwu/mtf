package framework

import (
	"testing"

	pb "github.com/smallinsky/mtf/e2e/proto/echo"
	pbo "github.com/smallinsky/mtf/e2e/proto/oracle"
	"github.com/smallinsky/mtf/framework"
	"github.com/smallinsky/mtf/port"
)

func TestMain(m *testing.M) {
	sutEnv := map[string]string{
		"ORACLE_ADDR": "host.docker.internal:8002",
	}
	framework.NewSuite(m).
		SUTEnv(sutEnv).
		SetMigratePath("../../e2e/migrations").
		SetSUTPath("/Users/marek/Go/src/github.com/smallinsky/mtf/e2e/service/echo/").
		Run()
}

func TestEchoService(t *testing.T) {
	framework.Run(t, new(SuiteTest))
}

func (st *SuiteTest) Init(t *testing.T) {
	var err error
	if st.echoPort, err = port.NewGRPCClientPort((*pb.EchoClient)(nil), "localhost:8001"); err != nil {
		t.Fatalf("failed to init grpc client port")
	}
	if st.httpPort, err = port.NewHTTPPort(port.WithTLSHost("*.icndb.com")); err != nil {
		t.Fatalf("failed to init http port")
	}
	if st.oraclePort, err = port.NewGRPCServerPort((*pbo.OracleServer)(nil), ":8002"); err != nil {
		t.Fatalf("failed to init grpc oracle server")
	}
}

type SuiteTest struct {
	echoPort   *port.Port
	httpPort   *port.Port
	oraclePort *port.Port
}

func (st *SuiteTest) TestRedis(t *testing.T) {
	st.echoPort.Send(t, &pb.AskRedisRequest{
		Data: "make me sandwitch",
	})
	st.echoPort.Receive(t, &pb.AskRedisResponse{
		Data: "what? make it yourself",
	})
	st.echoPort.Send(t, &pb.AskRedisRequest{
		Data: "sudo make me sandwitch",
	})
	st.echoPort.Receive(t, &pb.AskRedisResponse{
		Data: "okey",
	})
}

func (st *SuiteTest) TestHTTP(t *testing.T) {
	st.echoPort.Send(t, &pb.AskGoogleRequest{
		Data: "Get answer for ultimate question of life the universe and everything",
	})
	st.httpPort.Receive(t, &port.HTTPRequest{
		Method: "GET",
	})
	st.httpPort.Send(t, &port.HTTPResponse{
		Body: []byte(`{"value":{"joke":"42"}}`),
	})
	st.echoPort.Receive(t, &pb.AskGoogleResponse{
		Data: "42",
	})
}

func (st *SuiteTest) TestClientServerGRPC(t *testing.T) {
	st.echoPort.Send(t, &pb.AskOracleRequest{
		Data: "Get answer for ultimate question of life the universe and everything",
	})
	st.oraclePort.Receive(t, &pbo.AskDeepThroughRequest{
		Data: "Get answer for ultimate question of life the universe and everything",
	})
	st.oraclePort.Send(t, &pbo.AskDeepThroughRespnse{
		Data: "42",
	})
	st.echoPort.Receive(t, &pb.AskOracleResponse{
		Data: "42",
	})
}

func (st *SuiteTest) TestFetchDataFromDB(t *testing.T) {
	st.echoPort.Send(t, &pb.AskDBRequest{
		Data: "the dirty fork",
	})
	st.echoPort.Receive(t, &pb.AskDBResponse{
		Data: "Lucky we didn't say anything about the dirty knife",
	})
}
