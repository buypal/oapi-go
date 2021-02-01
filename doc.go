// Package oapi will convert your go types to schemes allowing to make
// go code parital source of truth.
//
// With many openapi libraries around there is common patter to keep
// your documentation in comments. It is usually done with some 'magic'
// syntax which can be utilized to make your program to be source of truth
// for openapi specifications.
//
// Problem with this approach is that if you want to do specification well and allow
// more complicated use cases you are forced to use yaml embedded in comments.
//
// I would argue that we are better off with using just good ol' plain yaml files.
// Defining small openapi spec in yaml files is easy, tooling works well, and its flexible
// up to the point where you have to manage large amount of data structures.
//
// This package is aming to provide tools which helps to manage situation
// where simple yaml files are burden rather then simple solution.
//
// This is done by making use of json pointers, which are in core of
// json schema and openapi. More about json pointers:
// https://tools.ietf.org/html/draft-ietf-appsawg-json-pointer-04
//
// Go source and its structure can be navigated with json pointers
// same way yaml or json can be. Syntax of json pointer is bit altered but
// it is still compliant with RFC.
//
// Go Pointer resolution
//
// Given go structure:
//  type Object struct { Field string `json:"field"` }
// And defined specification:
//  {"$ref": "go://github.com/buypal/oapi-go#/Object" }
// Following json scheme will be produced:
//  {"schema": {"properties": {"field": {"type": "string"}}}}
//
// Pointer go://github.com/buypal/oapi-go#/Object uri refers to
// package and fragment (pointer) is referring to exact location of
// data structure in that package.
//
// Following pointers will be resolved as json schema:
//
// • go://github.com/buypal/oapi-go#/Object
//
// • go://github.com/buypal/oapi-go#/Object/Field
//
// • go://#/Object (local resolution)
//
// With this approach we can tackle problem of having resistance against changes
// in go source code. For example if you change data type of Field from `string` to `int`
// json schema will become:
//  {"schema": {"properties": {"field": {"type": "number", "format": "int32"}}}}
//
// By default all schemes are resolved locally. If you want export scheme globally
// you can use magic comments in your go source code.
//
//  //oapi:schema <schema name> <source of schema>
//  //oapi:schema <source of schema>
//
// Examples (works same):
//  //oapi:schema Object
//  //oapi:schema go://github.com/buypal/oapi-go#/Object
//
// Merging specifications
//
// One of the goals of this package was also to provide way how to merge multiple
// specifications together. Every package can have file oapi.yaml or oapi.json which will
// be merged to "global" document. This is happening out of the box just
// by running oapi command.
// This allows to mantain per package openapi specifications.
//
// Additional RFC documents
//
// https://tools.ietf.org/html/rfc3986
//
// https://tools.ietf.org/html/draft-ietf-appsawg-json-pointer-04
//
// https://tools.ietf.org/id/draft-pbryan-zyp-json-ref-03.html#RFC3986
package oapi
