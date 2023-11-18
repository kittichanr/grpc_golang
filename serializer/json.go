package serializer

import (
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func ProtobufToJson(message proto.Message) (string, error) {
	marshaler := protojson.MarshalOptions{
		UseEnumNumbers:  false,
		EmitUnpopulated: true,
		Indent:          " ",
		UseProtoNames:   true,
	}
	jsonBytes, err := marshaler.Marshal(message)
	return string(jsonBytes), err
}
