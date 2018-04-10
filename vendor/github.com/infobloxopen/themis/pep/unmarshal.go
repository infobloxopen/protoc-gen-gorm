package pep

import (
	"errors"
	"fmt"
	"math"
	"net"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/infobloxopen/themis/pdp"
	pb "github.com/infobloxopen/themis/pdp-service"
)

const (
	effectFieldName = "Effect"
	reasonFieldName = "Reason"
)

var (
	// ErrorInvalidDestination indicates that output value of validate method is
	// not a structure.
	ErrorInvalidDestination = errors.New("given value is not a pointer to structure")
)

type resFieldsInfo struct {
	fields map[string]string
	err    error
}

var (
	resTypeCache     = map[string]resFieldsInfo{}
	resTypeCacheLock = sync.RWMutex{}
)

func fillResponse(res *pb.Response, v interface{}) error {
	if out, ok := v.(*pb.Response); ok {
		*out = *res
		return nil
	}

	return unmarshalToValue(res, reflect.ValueOf(v))
}

func unmarshalToValue(res *pb.Response, v reflect.Value) error {
	if v.Kind() != reflect.Ptr {
		return ErrorInvalidDestination
	}

	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return ErrorInvalidDestination
	}

	fields, err := makeFieldMap(v.Type())
	if err != nil {
		return err
	}

	if len(fields) > 0 {
		return unmarshalToTaggedStruct(res, v, fields)
	}

	return unmarshalToUntaggedStruct(res, v)
}

func parseTag(tag string, f reflect.StructField, t reflect.Type) (string, error) {
	items := strings.Split(tag, ",")
	if len(items) > 1 {
		tag = items[0]
		taggedTypeName := items[1]

		if tag == effectFieldName || tag == reasonFieldName {
			return "", fmt.Errorf("don't support type definition for \"%s\" and \"%s\" fields (%s.%s)",
				effectFieldName, reasonFieldName, t.Name(), f.Name)
		}

		taggedType, ok := typeByTag[strings.ToLower(taggedTypeName)]
		if !ok {
			return "", fmt.Errorf("unknown type \"%s\" (%s.%s)", taggedTypeName, t.Name(), f.Name)
		}

		if taggedType != f.Type {
			return "", fmt.Errorf("tagged type \"%s\" doesn't match field type \"%s\" (%s.%s)",
				taggedTypeName, f.Type.Name(), t.Name(), f.Name)
		}

		return tag, nil
	}

	if tag == effectFieldName {
		return effectFieldName, nil
	}

	if tag == reasonFieldName {
		return reasonFieldName, nil
	}

	return tag, nil
}

func makeFieldMap(t reflect.Type) (map[string]string, error) {
	key := t.PkgPath() + "." + t.Name()
	resTypeCacheLock.RLock()
	if info, ok := resTypeCache[key]; ok {
		resTypeCacheLock.RUnlock()
		return info.fields, info.err
	}
	resTypeCacheLock.RUnlock()

	m := make(map[string]string)
	var err error
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		tag, ok := getTag(f)
		if !ok {
			continue
		}

		if len(tag) <= 0 {
			tag, ok = getName(f)
			if !ok {
				continue
			}
		}

		tag, err = parseTag(tag, f, t)
		if err != nil {
			break
		}

		m[tag] = f.Name
	}

	resTypeCacheLock.Lock()
	resTypeCache[key] = resFieldsInfo{
		fields: m,
		err:    err,
	}
	resTypeCacheLock.Unlock()

	return m, err
}

type fieldUnmarshaller func(attr *pb.Attribute, v reflect.Value) error

var unmarshallersByType = map[string]fieldUnmarshaller{
	pdp.TypeKeys[pdp.TypeBoolean]: boolUnmarshaller,
	pdp.TypeKeys[pdp.TypeString]:  stringUnmarshaller,
	pdp.TypeKeys[pdp.TypeInteger]: intUnmarshaller,
	pdp.TypeKeys[pdp.TypeFloat]:   floatUnmarshaller,
	pdp.TypeKeys[pdp.TypeAddress]: addressUnmarshaller,
	pdp.TypeKeys[pdp.TypeNetwork]: networkUnmarshaller,
	pdp.TypeKeys[pdp.TypeDomain]:  domainUnmarshaller}

func unmarshalToTaggedStruct(res *pb.Response, v reflect.Value, fields map[string]string) error {
	name, ok := fields[effectFieldName]
	if ok {
		setToUntaggedEffect(res, v, name)
	}

	name, ok = fields[reasonFieldName]
	if ok {
		setToUntaggedReason(res, v, name)
	}

	for _, attr := range res.Obligation {
		name, ok := fields[attr.Id]
		if !ok {
			continue
		}

		f := v.FieldByName(name)
		if !f.CanSet() {
			return fmt.Errorf("field %s.%s is tagged but can't be set", v.Type().Name(), name)
		}

		unmarshaller, ok := unmarshallersByType[attr.Type]
		if !ok {
			return fmt.Errorf("can't unmarshal \"%s\" of \"%s\" type", attr.Id, attr.Type)
		}

		if t, ok := typeByTag[attr.Type]; ok {
			if t != f.Type() {
				return fmt.Errorf("can't unmarshal \"%s\" of \"%s\" type to field %s.%s",
					attr.Id, attr.Type, v.Type().Name(), name)
			}
		} else {
			return fmt.Errorf("can't unmarshal \"%s\" of \"%s\" type", attr.Id, attr.Type)
		}

		err := unmarshaller(attr, f)
		if err != nil {
			return err
		}
	}

	return nil
}

func setToUntaggedEffect(res *pb.Response, v reflect.Value, name string) bool {
	f := v.FieldByName(name)
	if !f.CanSet() {
		return false
	}

	k := f.Kind()
	if k == reflect.Bool {
		f.SetBool(res.Effect == pb.Response_PERMIT)
		return true
	}

	if k == reflect.Int || k == reflect.Int8 || k == reflect.Int16 || k == reflect.Int32 || k == reflect.Int64 {
		f.SetInt(int64(res.Effect))
		return true
	}

	if k == reflect.Uint || k == reflect.Uint8 || k == reflect.Uint16 || k == reflect.Uint32 || k == reflect.Uint64 {
		f.SetUint(uint64(res.Effect))
		return true
	}

	if k == reflect.String {
		f.SetString(pb.Response_Effect_name[int32(res.Effect)])
		return true
	}

	return false
}

func setToUntaggedReason(res *pb.Response, v reflect.Value, name string) bool {
	f := v.FieldByName(name)
	if !f.CanSet() {
		return false
	}

	if f.Kind() == reflect.String {
		f.SetString(res.Reason)
		return true
	}

	return false
}

func unmarshalToUntaggedStruct(res *pb.Response, v reflect.Value) error {
	skipEffect := setToUntaggedEffect(res, v, effectFieldName)
	skipReason := setToUntaggedReason(res, v, reasonFieldName)

	for _, attr := range res.Obligation {
		if attr.Id == effectFieldName && skipEffect {
			continue
		}

		if attr.Id == reasonFieldName && skipReason {
			continue
		}

		f := v.FieldByName(attr.Id)
		if !f.CanSet() {
			continue
		}

		unmarshaller, ok := unmarshallersByType[attr.Type]
		if !ok {
			return fmt.Errorf("can't unmarshal \"%s\" of \"%s\" type", attr.Id, attr.Type)
		}

		if t, ok := typeByTag[attr.Type]; !ok || t != f.Type() {
			continue
		}

		err := unmarshaller(attr, f)
		if err != nil {
			return err
		}
	}

	return nil
}

func boolUnmarshaller(attr *pb.Attribute, v reflect.Value) error {
	b, err := strconv.ParseBool(attr.Value)
	if err != nil {
		return fmt.Errorf("can't treat \"%s\" value (%s) as boolean: %s", attr.Id, attr.Value, err)
	}

	v.SetBool(b)
	return nil
}

func stringUnmarshaller(attr *pb.Attribute, v reflect.Value) error {
	v.SetString(attr.Value)
	return nil
}

func intUnmarshaller(attr *pb.Attribute, v reflect.Value) error {
	i, err := strconv.ParseInt(attr.Value, 0, 64)
	if err != nil {
		return fmt.Errorf("can't treat \"%s\" value (%s) as integer: %s", attr.Id, attr.Value, err)
	}

	switch v.Kind() {
	case reflect.Int:
		if i < math.MinInt32 || i > math.MaxInt32 {
			return fmt.Errorf("\"%s\" %d overflows int value", attr.Id, i)
		}

		v.SetInt(i)
		return nil

	case reflect.Int8:
		if i < math.MinInt8 || i > math.MaxInt8 {
			return fmt.Errorf("\"%s\" %d overflows int8 value", attr.Id, i)
		}

		v.SetInt(i)
		return nil

	case reflect.Int16:
		if i < math.MinInt16 || i > math.MaxInt16 {
			return fmt.Errorf("\"%s\" %d overflows int16 value", attr.Id, i)
		}

		v.SetInt(i)
		return nil

	case reflect.Int32:
		if i < math.MinInt32 || i > math.MaxInt32 {
			return fmt.Errorf("\"%s\" %d overflows int32 value", attr.Id, i)
		}

		v.SetInt(i)
		return nil

	case reflect.Int64:
		v.SetInt(i)
		return nil

	case reflect.Uint:
		if i < 0 || i > math.MaxUint32 {
			return fmt.Errorf("\"%s\" %d overflows uint value", attr.Id, i)
		}

		v.SetUint(uint64(i))
		return nil

	case reflect.Uint8:
		if i < 0 || i > math.MaxUint8 {
			return fmt.Errorf("\"%s\" %d overflows uint8 value", attr.Id, i)
		}

		v.SetUint(uint64(i))
		return nil

	case reflect.Uint16:
		if i < 0 || i > math.MaxUint16 {
			return fmt.Errorf("\"%s\" %d overflows uint16 value", attr.Id, i)
		}

		v.SetUint(uint64(i))
		return nil

	case reflect.Uint32:
		if i < 0 || i > math.MaxUint32 {
			return fmt.Errorf("\"%s\" %d overflows uint32 value", attr.Id, i)
		}

		v.SetUint(uint64(i))
		return nil

	case reflect.Uint64:
		if i < 0 {
			return fmt.Errorf("\"%s\" %d overflows uint64 value", attr.Id, i)
		}

		v.SetUint(uint64(i))
		return nil

	}

	return fmt.Errorf("can't set value %q of \"%s\" attribute to %s", attr.Value, attr.Id, v.Type().Name())
}

func floatUnmarshaller(attr *pb.Attribute, v reflect.Value) error {
	f, err := strconv.ParseFloat(attr.Value, 64)
	if err != nil {
		return fmt.Errorf("can't treat \"%s\" value (%s) as integer: %s", attr.Id, attr.Value, err)
	}

	switch v.Kind() {
	case reflect.Float32:
		absF := math.Abs(f)
		if absF > math.MaxFloat32 {
			return fmt.Errorf("\"%s\" %g overflows float32 value", attr.Id, f)
		}
		if absF < math.SmallestNonzeroFloat32 {
			return fmt.Errorf("\"%s\" %g underflows float32 value", attr.Id, f)
		}
		v.SetFloat(f)
		return nil

	case reflect.Float64:
		v.SetFloat(f)
		return nil
	}

	return fmt.Errorf("can't set value %q of \"%s\" attribute to %s", attr.Value, attr.Id, v.Type().Name())
}

func addressUnmarshaller(attr *pb.Attribute, v reflect.Value) error {
	s := attr.Value
	if strings.Contains(s, ":") {
		if strings.Contains(s, "]") {
			s = strings.Split(s, "]")[0][1:]
		} else if strings.Contains(s, ".") {
			s = strings.Split(s, ":")[0]
		}
	}

	ip := net.ParseIP(s)
	if ip == nil {
		return fmt.Errorf("can't treat \"%s\" value (%s) as address", attr.Id, attr.Value)
	}

	v.Set(reflect.ValueOf(ip))
	return nil
}

func networkUnmarshaller(attr *pb.Attribute, v reflect.Value) error {
	_, n, err := net.ParseCIDR(attr.Value)
	if err != nil {
		return fmt.Errorf("can't treat \"%s\" value (%s) as network: %s", attr.Id, attr.Value, err)
	}

	v.Set(reflect.ValueOf(*n))
	return nil
}

func domainUnmarshaller(attr *pb.Attribute, v reflect.Value) error {
	v.SetString(attr.Value)
	return nil
}
