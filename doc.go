// Package oapi will convert your go types to schemes allowing to make
// go source parital source of truth.
//
// With many openapi libraries around there is common patter to keep
// your documentation in comments. It is usually done with some 'magic'
// syntax which can be utilized to make your program to be source of truth
// for openapi specifications.
//
// Problem with this approach is that if you want to do specification well and allow
// more complicated usecases you are forced to use yaml embedded in comments.
//
// At this point I would argue that we are better of with using just good ol' plain yaml.
// Defining small openapi spec in yaml is easy, tooling works well, and its flexible
// up to the point where you have to manage large amount of strcutures
// and its hard to keep track of what is what.
// At this point you might find oapi package useful.
//
// Oapi package is aming to make use of json poainters, which are possible to
// be used by openapi specification as well. More about json pointers:
// https://tools.ietf.org/html/draft-ietf-appsawg-json-pointer-04
//
// Go source and its structure can be navigated with json pointers
// same way yaml or json can be. Syntax of json pointer is bit altered but
// it is still complient with RFC.
//
// Given go strcuture:
//  type Object struct { Field string `json:"field"` }
// And defined specifiction:
//  {"$ref": "go://github.com/buypal/oapi-go#/Object" }
// Following json scheme will be produced:
//  {"schema": {"properties": {"field": {"type": "string"}}}}
//
// Pointer go://github.com/buypal/oapi-go#/Object uri referes to
// package and fragment (pointer) is refereing to excat location of
// data strcuture in that package.
//
// Following pointers will be resolved as json schema:
// - go://github.com/buypal/oapi-go#/Object
// - go://github.com/buypal/oapi-go#/Object/Field
// - go://#/Object (local resolution)
//
// With this approach we can tackle problem of having restistency againts changes
// in go source code. For example if you change data type of Field from `string` to `int`
// json schema will become:
//  {"schema": {"properties": {"field": {"type": "number", "format": "int32"}}}}
//
// By default all schemes are resolved locally. If you want export scheme globally
// you can use magic comment in your go source code.
//
// //oapi:schema <schema name> <source of schema>
// //oapi:schema <source of schema>
//
// Examples (both means same):
// //oapi:schema Object
// //oapi:schema go://github.com/buypal/oapi-go#/Object
//
// This package will also allow you to merge multiple specifications together.
// In nuthsell every go package can have its own specification wich will be merged
// into single one. This allows to mantain per package openapi specifications.
// In your go package you can define oapi.yaml and as long as you execute oapi
// and given package will be maked as go dependency import it will be processed na merged.
//
// Additional RFC documents:
// - https://tools.ietf.org/html/rfc3986
// - https://tools.ietf.org/html/draft-ietf-appsawg-json-pointer-04
// - https://tools.ietf.org/id/draft-pbryan-zyp-json-ref-03.html#RFC3986
package oapi
