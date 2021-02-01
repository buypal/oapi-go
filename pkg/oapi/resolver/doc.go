package resolver

// interface and marshal problem - can be marshaled to anything need to have $ref oneof allof anyof etc..
// need to be able override stdlib, time.Time example
// constants
// tag decoration for types which are primitives not structures

//oapi:schema url
//oapi:response url
//oapi:tag url `hello,type:'number',allof:'go:pkg/timetime#/Time, '`

// 1 or more overrides configs
// first export scheme in go and later you can resolve given scheme with go url, keep in mind
// referencing in yaml non existing scheme which is exported in go but referenced in yaml as #/components/scheme/reference
// is not an valid option unless #/components/scheme/reference exists in yaml first.

// local scope is resolved to full url always
// overrides in config are resolved as first
