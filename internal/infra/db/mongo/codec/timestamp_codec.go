package codec

import (
	"fmt"
	"reflect"

	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	tTimestampPb = reflect.TypeOf((*timestamppb.Timestamp)(nil))
)

// TimestampCodec handles encoding/decoding of timestamppb.Timestamp to/from BSON DateTime
type TimestampCodec struct{}

// EncodeValue converts timestamppb.Timestamp to BSON DateTime
func (tc *TimestampCodec) EncodeValue(ec bsoncodec.EncodeContext, vw bsonrw.ValueWriter, val reflect.Value) error {
	if !val.IsValid() || val.IsNil() {
		return vw.WriteNull()
	}

	timestamp, ok := val.Interface().(*timestamppb.Timestamp)
	if !ok {
		return fmt.Errorf("expected *timestamppb.Timestamp, got %T", val.Interface())
	}

	// Convert protobuf timestamp to Go time.Time, then to BSON DateTime
	t := timestamp.AsTime()
	dt := primitive.NewDateTimeFromTime(t)
	return vw.WriteDateTime(int64(dt))
}

// DecodeValue converts BSON DateTime or embedded document to timestamppb.Timestamp
func (tc *TimestampCodec) DecodeValue(dc bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value) error {
	if !val.CanSet() {
		return fmt.Errorf("cannot set value")
	}

	// Get the current BSON type
	bsonType := vr.Type()

	// Handle null values
	if bsonType == bsontype.Null {
		err := vr.ReadNull()
		if err != nil {
			return err
		}
		val.Set(reflect.Zero(val.Type()))
		return nil
	}

	// Handle BSON DateTime (from Python datetime or direct BSON DateTime)
	if bsonType == bsontype.DateTime {
		// Read BSON DateTime (int64 milliseconds since epoch)
		dt, err := vr.ReadDateTime()
		if err != nil {
			return fmt.Errorf("failed to read DateTime: %w", err)
		}

		// Convert BSON DateTime to Go time.Time
		t := primitive.DateTime(dt).Time()

		// Convert to timestamppb.Timestamp
		timestamp := timestamppb.New(t)

		// Set the value
		val.Set(reflect.ValueOf(timestamp))
		return nil
	}

	// Handle embedded document (protobuf format with seconds/nanos fields)
	if bsonType == bsontype.EmbeddedDocument {
		// Read the embedded document field by field
		dr, err := vr.ReadDocument()
		if err != nil {
			return fmt.Errorf("failed to read embedded document: %w", err)
		}

		var seconds int64
		var nanos int32

		// Read each field from the embedded document
		for {
			key, vr, err := dr.ReadElement()
			if err == bsonrw.ErrEOD {
				break
			}
			if err != nil {
				return fmt.Errorf("failed to read element: %w", err)
			}

			switch key {
			case "seconds":
				seconds, err = vr.ReadInt64()
				if err != nil {
					return fmt.Errorf("failed to read seconds: %w", err)
				}
			case "nanos":
				nanos32, err := vr.ReadInt32()
				if err != nil {
					return fmt.Errorf("failed to read nanos: %w", err)
				}
				nanos = nanos32
			default:
				// Skip unknown fields
				err = vr.Skip()
				if err != nil {
					return fmt.Errorf("failed to skip field %s: %w", key, err)
				}
			}
		}

		// Create timestamppb.Timestamp from seconds and nanos
		timestamp := &timestamppb.Timestamp{
			Seconds: seconds,
			Nanos:   nanos,
		}

		// Set the value
		val.Set(reflect.ValueOf(timestamp))
		return nil
	}

	return fmt.Errorf("expected DateTime or embedded document, got %v", bsonType)
}

// GetRegistry creates a new BSON codec registry with timestamppb.Timestamp support
func GetRegistry() *bsoncodec.Registry {
	// Start with MongoDB's default registry
	rb := bsoncodec.NewRegistryBuilder()

	// Add default codecs from MongoDB driver
	bsoncodec.DefaultValueEncoders{}.RegisterDefaultEncoders(rb)
	bsoncodec.DefaultValueDecoders{}.RegisterDefaultDecoders(rb)

	// Register the custom codec for *timestamppb.Timestamp
	rb.RegisterTypeEncoder(tTimestampPb, &TimestampCodec{})
	rb.RegisterTypeDecoder(tTimestampPb, &TimestampCodec{})

	return rb.Build()
}
