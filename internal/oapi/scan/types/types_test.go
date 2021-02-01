package types

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"strings"
	"testing"

	"github.com/buypal/oapi-go/internal/container"
	"github.com/buypal/oapi-go/internal/pointer"
	"github.com/buypal/oapi-go/tag"
	"github.com/stretchr/testify/require"
)

func pkgFor(source string, info *types.Info) (*types.Package, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", source, 0)
	if err != nil {
		return nil, err
	}
	conf := types.Config{Importer: importer.Default()}
	return conf.Check(f.Name.Name, fset, []*ast.File{f}, info)
}

func mustTypecheck(t *testing.T, source string, info *types.Info) {
	pkg, err := pkgFor(source, info)
	if err != nil {
		var name string
		if pkg != nil {
			name = "package " + pkg.Name()
		}
		t.Fatalf("%s: didn't type-check (%s)", name, err)
	}
}

func compileType(t *testing.T, what string, src string) types.Object {
	info := types.Info{
		Scopes: make(map[ast.Node]*types.Scope),
	}
	src = fmt.Sprintf("package %s\n%v", "test", src)
	mustTypecheck(t, src, &info)
	for _, s := range info.Scopes {
		f := s.Parent().Lookup(what)
		if f != nil {
			return f
		}
	}
	t.Fatalf("could not found type %q", what)
	return nil
}

func mustPoint(t *testing.T, name string) pointer.Pointer {
	x, err := pointer.NewGoPointer("test", name)
	require.NoError(t, err)
	return x
}

func TestCollectTypes(t *testing.T) {
	tp := compileType(t, "test", `
		type test struct {
			A string
			b string
		}
	`)
	var m pointmap
	require.NoError(t, collectTypes(tp, &m))

	tp2, ok := m.findType(mustPoint(t, "test"))
	require.True(t, ok)
	require.True(t, types.Identical(tp.Type().Underlying(), tp2.Underlying()))

	tp3, ok := m.findType(mustPoint(t, "test/A"))
	require.True(t, ok)
	require.Equal(t, tp3.String(), "string")

	_, ok = m.findType(mustPoint(t, "test/b"))
	require.False(t, ok)

	_, ok = m.findType(mustPoint(t, "test3"))
	require.False(t, ok)
}

func TestCollectTypesDeep(t *testing.T) {
	tp := compileType(t, "test1", `
		type test1 struct {
			A string
			B test2
		}
		type test2 struct {
			A string
		}
	`)
	var m pointmap
	require.NoError(t, collectTypes(tp, &m))

	tp1, ok := m.findType(mustPoint(t, "test1"))
	require.True(t, ok)
	require.Equal(t, tp1.Underlying().String(), "struct{A string; B test.test2}")

	tp2, ok := m.findType(mustPoint(t, "test1/A"))
	require.True(t, ok)
	require.Equal(t, tp2.Underlying().String(), "string")

	tp3, ok := m.findType(mustPoint(t, "test1/B"))
	require.True(t, ok)
	require.Equal(t, tp3.Underlying().String(), "struct{A string}")

	tp4, ok := m.findType(mustPoint(t, "test2/A"))
	require.True(t, ok)
	require.Equal(t, tp4.Underlying().String(), "string")
}

func TestCollectTypesPointer(t *testing.T) {
	tp := compileType(t, "test1", `
		type test1 *string // yeah this is a valid type
	`)
	var m pointmap
	require.NoError(t, collectTypes(tp, &m))

	tp1, ok := m.findType(mustPoint(t, "test1"))
	require.True(t, ok)
	require.Equal(t, tp1.Underlying().String(), "*string")

	tp2, ok := m.findType(mustPoint(t, "test1/*"))
	require.True(t, ok)
	require.Equal(t, tp2.Underlying().String(), "string")
}

func TestCollectTypesSelfReferencing(t *testing.T) {
	tp := compileType(t, "test1", `
		type test1 *test1 // yeah this is a valid type
	`)
	var m pointmap
	require.NoError(t, collectTypes(tp, &m))

	tp1, ok := m.findType(mustPoint(t, "test1"))
	require.True(t, ok)
	require.Equal(t, tp1.Underlying().String(), "*test.test1")

	tp2, ok := m.findType(mustPoint(t, "test1/*"))
	require.True(t, ok)
	require.Equal(t, tp2.Underlying().String(), "*test.test1")
}

func TestCollectTypesMap(t *testing.T) {
	tp := compileType(t, "test1", `
		type test1 map[string]string 
	`)
	var m pointmap
	require.NoError(t, collectTypes(tp, &m))

	tp1, ok := m.findType(mustPoint(t, "test1"))
	require.True(t, ok)
	require.Equal(t, tp1.Underlying().String(), "map[string]string")

	tp2, ok := m.findType(mustPoint(t, "test1/[]"))
	require.True(t, ok)
	require.Equal(t, tp2.Underlying().String(), "string")
}

func TestT2S(t *testing.T) {
	testCases := []struct {
		desc   string
		expr   string
		schema string
	}{
		{
			expr:   `type test string `,
			schema: `type: string`,
		},
		{
			expr:   `type test uint `,
			schema: "type: integer\nminimum: 0\nformat: int32",
		},
		{
			expr:   `type test uint8 `,
			schema: "type: integer\nminimum: 0\nformat: int32",
		},
		{
			expr:   `type test uint16 `,
			schema: "type: integer\nminimum: 0\nformat: int32",
		},
		{
			expr:   `type test uint32 `,
			schema: "type: integer\nminimum: 0\nformat: int32",
		},
		{
			expr:   `type test uint64 `,
			schema: "type: integer\nminimum: 0\nformat: int64",
		},
		{
			expr:   `type test int `,
			schema: "type: integer\nformat: int32",
		},
		{
			expr:   `type test int8 `,
			schema: "type: integer\nformat: int32",
		},
		{
			expr:   `type test int16 `,
			schema: "type: integer\nformat: int32",
		},
		{
			expr:   `type test int32 `,
			schema: "type: integer\nformat: int32",
		},
		{
			expr:   `type test int64 `,
			schema: "type: integer\nformat: int64",
		},
		{
			expr:   `type test float32 `,
			schema: "type: number\nformat: float",
		},
		{
			expr:   `type test float64 `,
			schema: "type: number\nformat: double",
		},
		{
			expr:   `type test bool `,
			schema: "type: boolean",
		},
		{
			expr: `
type test map[string]string 
`,
			schema: `
additionalProperties:
  type: string
nullable: true
type: object
`,
		},
		{
			expr: `
type test struct {
	A string
}
`,
			schema: `
properties:
  A:
    type: string
type: object
`,
		},
		{
			expr: `
type test struct {
	A string ` + "`" + `oapi:"a"` + "`" + `
}
`,
			schema: `
properties:
  a:
    type: string
type: object
`,
		},
		{
			expr: `
type test struct {
	A string ` + "`" + `oapi:"-"` + "`" + `
}
`,
			schema: `
type: object
`,
		},
		{
			expr: `
type test struct {
	A string ` + "`" + `json:"b" oapi:"a"` + "`" + `
}
`,
			schema: `
properties:
  a:
    type: string
type: object
`,
		},
		{
			expr: `
type test struct {
	A struct {
		B string
	}
}
`,
			schema: `
properties:
  A:
    $ref: go://test#/test/A
type: object
`,
		},
		{
			expr: `
type test struct {
	A struct {
		B string
	} ` + "`" + `oapi:",inline"` + "`" + `
}
`,
			schema: `
properties:
  B:
    type: string
type: object
`,
		},
		{
			expr: `
type test struct {
	Embed
}
type Embed struct {
	B string
}
`,
			schema: `
properties:
  B:
    type: string
type: object
`,
		},
		{
			expr: `type test *string`,
			schema: `
type: string
nullable: true
`,
		},
		{
			expr: `type test []string`,
			schema: `
type: array
items:
  type: string
nullable: true
`,
		},
		{
			expr: `type test [5]string`,
			schema: `
type: array
items:
  type: string
maxItems: 5
`,
		},
		{
			expr: `
type test struct {
	A string ` + "`" + `oapi:"a,type:number"` + "`" + `
}
`,
			schema: `
properties:
  a:
    type: number
type: object
`,
		},
		{
			expr: `
import "time"

type test struct {
	A time.Time
}
`,
			schema: `
properties:
  A:
    type: string
type: object
`,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			expr := strings.Trim(tC.expr, "\n\t")
			schema := strings.Trim(tC.schema, "\n\t")

			tp := compileType(t, "test", expr)

			var m pointmap
			require.NoError(t, collectTypes(tp, &m))
			sp, err := type2schema(tp.Type(), m, path{}, tag.Tag{})
			require.NoError(t, err)

			c1, err := container.ReadYAML([]byte(schema))
			require.NoError(t, err)

			c2, err := container.Make(sp)
			require.NoError(t, err)

			e1, _ := c1.MarshalYAML()
			e2, _ := c2.MarshalYAML()
			require.Equal(t, string(e1), string(e2))
		})
	}
}
