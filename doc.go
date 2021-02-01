// Package oapigo implementing functions which will help to
// parse go source code and find structures and syntax related to oapi.
// URLs for lawyers:
// - https://tools.ietf.org/html/rfc3986
// - https://tools.ietf.org/html/draft-ietf-appsawg-json-pointer-04
// - https://tools.ietf.org/id/draft-pbryan-zyp-json-ref-03.html#RFC3986
//
// In your go code you can use following syntax
//
//oapi:schema <schema name> <source of schema>
//
// Following URL addresses will be resolved as $ref in yaml:
// - http://github.com/buypal/oapi-go/schema.yaml
// - http://github.com/buypal/oapi-go/schema.yaml#/Object
// - file://schema.yaml#/Object
// - ./schema.yaml#/Object
// - go://github.com/buypal/oapi-go
// - go://github.com/buypal/oapi-go#/Object
// - go#/Object (current pkg)
//
// Here are som examples:
//
//oapi:namespace somepkg
//
//oapi:merge file://ok.yaml
//oapi:merge ./schema.yaml
//
// type Something struct {}
//
//oapi:schema Something
//oapi:schema Something go://github.com/buypal/oapi-go#/Something
//
//
// usage: 'go:generate oapigo oapi.yaml'
// usage: 'go:generate oapigo oapi.yaml --include="**/**.yaml"'
//
package oapigo
