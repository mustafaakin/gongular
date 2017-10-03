package gongular

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/asaskevich/govalidator"
)

const (
	// PlaceParameter is used in ValidationError to indicate the error is in
	// URL Parameters
	PlaceParameter = "URL Path Parameter"
	// PlaceQuery is used in ValidationError to indicate the error is in
	// Query parameters
	PlaceQuery = "Query Parameter"
	// PlaceBody is used in ValidationError to indicate the error is in
	// Body of the request
	PlaceBody = "Body"
	// PlaceForm is used in ValidationError to indicate the error is in
	// submitted form
	PlaceForm = "Form Value"
)

const (
	// FieldParameter defines the struct field name for looking up URL Parameters
	FieldParameter = "Param"
	// FieldBody defines the struct field name for looking up the body of request
	FieldBody = "Body"
	// FieldForm defines the struct field name for looking up form of request
	FieldForm = "Form"
	// FieldQuery defines the struct field name for looking up QUery Parameters
	FieldQuery = "Query"
)

const (
	// TagInject The field name that is used to lookup injections in the handlers
	TagInject = "inject"
)

func parseInt(kind reflect.Kind, s string, place string, field reflect.StructField, val *reflect.Value) error {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return ParseError{
			Place:     place,
			FieldName: field.Name,
			Reason:    fmt.Sprintf("The '%s' is not parseable to a integer", s),
		}
	}

	ok, lower, upper := checkIntRange(kind, i)
	if !ok {
		return ParseError{
			Place:     place,
			FieldName: field.Name,
			Reason:    fmt.Sprintf("Supplied value %d is not in range [%d, %d]", i, lower, upper),
		}
	}

	val.SetInt(i)
	return nil
}

func parseUint(kind reflect.Kind, s string, place string, field reflect.StructField, val *reflect.Value) error {
	i, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return ParseError{
			Place:     place,
			FieldName: field.Name,
			Reason:    fmt.Sprintf("The '%s' is not parseable to int", s),
		}
	}

	ok, lower, upper := checkUIntRange(kind, i)
	if !ok {
		return ParseError{
			Place:     place,
			FieldName: field.Name,
			Reason:    fmt.Sprintf("Supplied value %d is not in range [%d, %d]", i, lower, upper),
		}
	}

	val.SetUint(i)
	return nil
}

func parseFloat(kind reflect.Kind, s string, place string, field reflect.StructField, val *reflect.Value) error {
	i, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return ParseError{
			Place:     place,
			FieldName: field.Name,
			Reason:    fmt.Sprintf("The '%s' is not parseable to float/double", s),
		}
	}

	ok, lower, upper := checkFloatRange(kind, i)
	if !ok {
		return ParseError{
			Place:     place,
			FieldName: field.Name,
			Reason:    fmt.Sprintf("Supplied value %f is not in range [%f, %f]", i, lower, upper),
		}
	}

	val.SetFloat(i)
	return nil
}

func parseBool(s string, place string, field reflect.StructField, val *reflect.Value) error {
	switch strings.ToLower(s) {
	case "true", "1", "yes":
		val.SetBool(true)
	case "false", "0", "no":
		val.SetBool(false)
	default:
		return ParseError{
			FieldName: field.Name,
			Place:     place,
			Reason:    fmt.Sprintf("The '%s' is not a boolean", s),
		}
	}
	return nil
}

func parseSimpleParam(s string, place string, field reflect.StructField, val *reflect.Value) error {
	kind := field.Type.Kind()
	var err error
	switch kind {
	case reflect.String:
		val.SetString(s)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		err = parseInt(kind, s, place, field, val)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		err = parseUint(kind, s, place, field, val)
	case reflect.Float32, reflect.Float64:
		err = parseFloat(kind, s, place, field, val)
	case reflect.Bool:
		err = parseBool(s, place, field, val)
	}
	return err
}

func validateStruct(obj reflect.Value, place string) error {
	isValid, err := govalidator.ValidateStruct(obj.Interface())
	if !isValid {
		m := govalidator.ErrorsByField(err)
		return ValidationError{
			Place:  place,
			Fields: m,
		}
	}
	return nil
}

func (c *Context) parseParams(obj reflect.Value) error {
	param := obj.FieldByName(FieldParameter)
	paramType := param.Type()

	numFields := paramType.NumField()
	for i := 0; i < numFields; i++ {
		field := paramType.Field(i)

		s := c.Params().ByName(field.Name)
		val := param.Field(i)
		err := parseSimpleParam(s, PlaceParameter, field, &val)
		if err != nil {
			return err
		}
	}

	return validateStruct(param, PlaceParameter)
}

func (c *Context) parseBody(handlerObject reflect.Value) error {
	// TODO: Cache body if possible?
	body := handlerObject.FieldByName(FieldBody)
	b := body.Addr().Interface()

	err := json.NewDecoder(c.Request().Body).Decode(b)
	if err != nil {
		return ParseError{
			Place:  PlaceBody,
			Reason: err.Error(),
		}
	}

	return validateStruct(body, PlaceBody)
}

func (c *Context) parseQuery(obj reflect.Value) error {
	query := obj.FieldByName(FieldQuery)
	queryType := query.Type()

	numFields := queryType.NumField()
	queryValues := c.Request().URL.Query()
	for i := 0; i < numFields; i++ {
		field := queryType.Field(i)

		s := queryValues.Get(field.Name)
		if s == "" {
			// Do not fail right now, it is the job of validator
			continue
		}

		val := query.Field(i)

		err := parseSimpleParam(s, PlaceQuery, field, &val)
		if err != nil {
			return err
		}
	}
	return validateStruct(query, PlaceQuery)
}

func (c *Context) parseForm(obj reflect.Value) error {
	form := obj.FieldByName(FieldForm)
	formType := form.Type()

	numFields := formType.NumField()

	for i := 0; i < numFields; i++ {
		field := formType.Field(i)
		// If it is a file, parse the form
		if field.Type == reflect.TypeOf(&UploadedFile{}) {
			file, header, err := c.Request().FormFile(field.Name)

			// TODO: Make it optional??
			if err == http.ErrMissingFile {
				return ParseError{
					Place:     PlaceForm,
					FieldName: field.Name,
					Reason:    "Was expecting a file, but could not found in the request.",
				}
			} else if err != nil {
				// It should be an internal error, therefore we do not wrap with ParseError
				return err
			}

			// Pack them to a single structure
			uploadedFile := &UploadedFile{
				File:   file,
				Header: header,
			}

			form.Field(i).Set(reflect.ValueOf(uploadedFile))
		} else {
			s := c.Request().FormValue(field.Name)
			val := form.Field(i)
			err := parseSimpleParam(s, PlaceForm, field, &val)
			if err != nil {
				return err
			}
		}
	}
	return validateStruct(form, PlaceForm)
}

func (c *Context) parseInjections(obj reflect.Value, injector *injector) error {
	numFields := obj.Type().NumField()

	for i := 0; i < numFields; i++ {
		field := obj.Type().Field(i)
		tip := field.Type
		name := field.Name

		// We can skip the field if it is a special one
		if name == FieldBody || name == FieldParameter || name == FieldQuery || name == FieldForm {
			continue
		}

		if !obj.Field(i).CanSet() {
			// It is an un-exported one
			continue
		}

		var key string
		tag, ok := field.Tag.Lookup(TagInject)
		if !ok {
			key = "default"
		} else {
			key = tag
		}
		fieldObj := obj.Field(i)

		err := c.setInjectionForField(tip, key, injector, fieldObj)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Context) setInjectionForField(tip reflect.Type, key string, injector *injector, fieldObj reflect.Value) error {
	cachedVal, cachedOk := c.getCachedInjection(tip, key)
	val, directOk := injector.GetDirectValue(tip, key)
	fn, customOk := injector.GetCustomValue(tip, key)

	if cachedOk {
		fieldObj.Set(reflect.ValueOf(cachedVal))
		return nil
	} else if directOk {
		fieldObj.Set(reflect.ValueOf(val))
		return nil
	} else if customOk {
		val, err := fn(c)
		if err != nil {
			return InjectionError{
				Key:             key,
				Tip:             tip,
				UnderlyingError: err,
			}
		}
		fieldObj.Set(reflect.ValueOf(val))
		c.putCachedInjection(tip, key, val)
		return nil
	}

	// We should not be here if the programmatic check is done correctly, but placed it here anyways
	return InjectionError{
		Key:             key,
		Tip:             tip,
		UnderlyingError: ErrNoSuchDependency,
	}
}
