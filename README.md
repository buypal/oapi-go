# OAPI GO

**Make your go application (partial) source of truth for your OpenAPI spec.**

This library acts as if go packages, yaml files, or json files were just json documents,
giving you ability to use pointers to reference and convert go types into schemes.


## Example

Given such a struct:

```go
package mypkg
// This package is located at github.com/org/repo/mypkg

type Item {
    Kind string `json:"kind"`
    Name string `json:"name"`
}
```

and such a definition:

```yaml
openapi: 3.0.3
info:
    title: "MyAwsomeApp"
    version: "-"
paths:
  /v1/items:
    get:
      responses:
        "200":
          content:
            application/json:
            schema:
              $ref: 'go://#/Item' # shorthand for 'go://github.com/org/repo/mypkg#/Item'
          description: Returns items
```

will produce OpenAPI specs:

```yaml
openapi: 3.0.3
info:
    title: "MyAwsomeApp"
    version: "-"
paths:
  /v1/items:
    get:
      responses:
        "200":
          content:
            application/json:
            schema:
              type: object
              properties:
                kind:
                  type: string
                name:
                  type: string
          description: Returns items
```

Of course this is very simple example just for illustration on how powerful pointers can be.

```
go://github.com/path/to/pkg#/Struct/Field
[1--][2--------------------][3----------]

For go pointers:

1) Is scheme, designating what resolver we are using in this case it is go.
2) Is full package path with authority.
3) Is path in go pkg to given struct field or any type.

Similary any protocol can be resolved this way (go://, file://, http://, etc).
```


## Motivation:

This library is trying to challange way how currently many openapi/swag libraries work.
Traditionally if you want to make your go application being source of truth of your 
swagger / openapi specification, you have to use "magic" comments which will be parsed
and specification will be generated.

Golang depends heavily on comments, tools like godoc, go generate, build flags would not 
work well without them. My personal believe is that screwing up docs or managing yaml
being pasted in comments is not worth it. 

I had following requirements:
- OpenAPI specs should be valid and used on go package basis 
- magic comments should be use but very lightly, rather as decoration
- go types should be source of truth and you should be able to reference them from any angle
- you should be able to merge multiple OpenAPI specs into one
- you should be able to support multiple specs in one repository
- ability to decorate structs with oapi tag to allow specifiying schema specific attributes
- ability to also have runtime validator based on oapi tag


## Usage:

```bash
usage: oapi [<flags>]

Flags:
  --help               Show context-sensitive help (also try --help-long and --help-man).
  --loglevel=LOGLEVEL  will set log level
  --config=CONFIG      config to be used
  --dir=DIR            execution directory usually dir of main pkg
  --format=FORMAT      will set output format
  --output=OUTPUT      will set output destination

```

Simpliest way to start is to define yaml definition:

```yaml
# in .examples/config/items/

openapi: "3.0.3"
info:
  title: "test"
  version: "-"
paths: 
  /v1/items:
    get:
      responses:
        "200":
          description: "Returns items"
          content:
            application/json:
              schema:
                $ref: 'go://#/Response'
```

And now you can run:

```bash
oapi --dir .examples/config/items/
```

output will be printed to stdout.

Note directory in this example needs to be pacakge importing all dependencies otherwise those 
will not be resolved.

We recommend to place config and `//go:generate` in your main package.


## Inspiration:

https://github.com/wzshiming/openapi

https://github.com/santhosh-tekuri/jsonschema