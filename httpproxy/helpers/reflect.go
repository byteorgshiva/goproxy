package helpers

import (
	"fmt"
	"net"
	"net/http"
	"reflect"
)

func ReflectRemoteAddrFromResponse(resp *http.Response) (net.Addr, error) {
	if resp.Body == nil {
		return nil, fmt.Errorf("ReflectRemoteAddrFromResponse: cannot reflect %#v for %v", resp, resp.Request.URL.String())
	}

	v := reflect.ValueOf(resp.Body)

	switch v.Type().String() {
	case "*http.bodyEOFSignal":
		v = v.Elem().FieldByName("body").Elem()
		v = reflect.Indirect(v).FieldByName("src").Elem()
		switch v.Type().String() {
		case "*io.LimitedReader":
			v = reflect.Indirect(v).FieldByName("R").Elem()
			v = reflect.Indirect(v).FieldByName("rd").Elem()
			v = reflect.Indirect(v).FieldByName("conn").Elem()
		default:
			return nil, fmt.Errorf("ReflectRemoteAddrFromResponse: unsupport %#v Type=%s", v, v.Type().String())
		}
	case "http2.transportResponseBody":
		v = v.FieldByName("cs").Elem()
		v = v.FieldByName("cc").Elem()
		v = v.FieldByName("tconn").Elem()
	default:
		return nil, fmt.Errorf("ReflectRemoteAddrFromResponse: unsupport %#v Type=%s", v, v.Type().String())
	}

	if v.Type().String() == "*tls.Conn" {
		v = reflect.Indirect(v).FieldByName("conn").Elem()
	}

	switch v.Type().String() {
	case "*net.TCPConn":
		v = reflect.Indirect(v).FieldByName("fd").Elem()
		v = reflect.Indirect(v).FieldByName("raddr").Elem()
		v1 := reflect.Indirect(reflect.Indirect(v).FieldByName("IP"))
		v2 := reflect.Indirect(reflect.Indirect(v).FieldByName("Port"))
		v3 := reflect.Indirect(reflect.Indirect(v).FieldByName("Zone"))
		return &net.TCPAddr{
			IP:   v1.Slice(0, v1.Len()).Bytes(),
			Port: int(v2.Int()),
			Zone: v3.String(),
		}, nil
	}

	return nil, fmt.Errorf("ReflectRemoteAddrFromResponse: unsupport %#v Type=%s", v, v.Type().String())
}
