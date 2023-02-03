package dut

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"testing"

	gnmipb "github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/ygnmi/ygnmi"
	"github.com/openconfig/ygot/ygot"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type GnmiClient struct {
	d           *DutApi
	location    string
	grpcConn    *grpc.ClientConn
	gnmiClient  *gnmipb.GNMIClient
	ygnmiClient *ygnmi.Client
}

type grpcPassCred struct {
	username string
	password string
}

func (c *grpcPassCred) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"username": c.username,
		"password": c.password,
	}, nil
}

func (c *grpcPassCred) RequireTransportSecurity() bool {
	return true
}

func NewGnmiClient(d *DutApi) (*GnmiClient, error) {
	location := fmt.Sprintf("%s:%d", d.dutConfig.Host, d.dutConfig.GnmiPort)
	d.t.Logf("Creating gNMI client for server %s ...\n", location)

	grpcConn, err := grpc.Dial(
		location,
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})),
		grpc.WithPerRPCCredentials(&grpcPassCred{
			username: d.dutConfig.GnmiUsername,
			password: d.dutConfig.GnmiPassword,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to dial gRPC for location %s: %v", location, err)
	}

	gnmiClient := gnmipb.NewGNMIClient(grpcConn)
	ygnmiClient, err := ygnmi.NewClient(gnmiClient, ygnmi.WithTarget(d.dutConfig.Name))
	if err != nil {
		return nil, fmt.Errorf("failed to create ygnmi client for location %s: %v", location, err)
	}

	return &GnmiClient{
		d:           d,
		location:    location,
		grpcConn:    grpcConn,
		gnmiClient:  &gnmiClient,
		ygnmiClient: ygnmiClient,
	}, nil
}

func DumpGnmiConfig(t *testing.T, msg string, s ygot.GoStruct) {
	m, err := ygot.ConstructIETFJSON(s, nil)
	if err != nil {
		t.Fatalf("Could not construct IETF JSON from YGOT struct: %v", err)
	}
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		t.Fatalf("Could not marshal IETF JSON: %v", err)
	}
	t.Logf("%s\n%s\n", msg, string(b))
}

// TODO: originally T was supposed to by 'any'
func GnmiReplace[T ygot.GoStruct](g *GnmiClient, q ygnmi.ConfigQuery[T], val T) {
	DumpGnmiConfig(g.d.t, "Pushing DUT config:", val)
	_, err := ygnmi.Replace(context.Background(), g.ygnmiClient, q, val)
	if err != nil {
		g.d.t.Fatal(err)
	}
}

// TODO: originally T was supposed to by 'any'
func GnmiGet[T ygot.GoStruct](g *GnmiClient, q ygnmi.SingletonQuery[T]) T {

	v, err := ygnmi.Get(context.Background(), g.ygnmiClient, q)
	if err != nil {
		g.d.t.Fatal(err)
	}
	DumpGnmiConfig(g.d.t, "Got DUT Config:", v)
	return v
}

// TODO: originally T was supposed to by 'any'
func GnmiDelete[T ygot.GoStruct](g *GnmiClient, q ygnmi.ConfigQuery[T]) {

	_, err := ygnmi.Delete(context.Background(), g.ygnmiClient, q)
	if err != nil {
		g.d.t.Fatal(err)
	}
}
