package volumegroup

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Response struct {
	Response interface{}
	Error    error
}

func (r *Response) HasKnownGRPCError(knownErrors []codes.Code) bool {
	if r.Error == nil {
		return false
	}

	s, ok := status.FromError(r.Error)
	if !ok {
		// This is not gRPC error. The operation must have failed before gRPC
		// method was called, otherwise we would get gRPC error.
		return false
	}

	for _, e := range knownErrors {
		if s.Code() == e {
			return true
		}
	}

	return false
}
