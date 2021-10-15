// Package cmd contains the primary logic of the xo command-line application.
package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/alecthomas/kingpin"
	"github.com/gobwas/glob"
	"github.com/xo/dburl"
	"github.com/xo/dburl/passfile"
	"github.com/calebhiebert/xo/loader"
	"github.com/calebhiebert/xo/models"
	xo "github.com/calebhiebert/xo/types"
	"os/user"
	"reflect"
)

func IntrospectSchema(ctx context.Context, schema, driver string, db *sql.DB) (*xo.XO, error) {
	ctx, err := SetupDatabase(ctx, schema, driver, db)
	if err != nil {
		return nil, err
	}

	ctx, args, err := NewArgs(ctx, &Args{
		Verbose: true,
		DbParams: DbParams{
			Schema: schema,
		},
		SchemaParams: SchemaParams{
			FkMode: "smart",
		},
	})
	if err != nil {
		return nil, err
	}

	f := BuildSchema

	x := new(xo.XO)

	if err := f(ctx, args, x); err != nil {
		return nil, err
	}

	return x, nil
}

func SetupDatabase(ctx context.Context, schema, driver string, db *sql.DB) (context.Context, error) {
	// add driver to context
	ctx = context.WithValue(ctx, xo.DriverKey, driver)

	// add db to context
	ctx = context.WithValue(ctx, xo.DbKey, db)

	var err error

	// determine schema
	if schema == "" {
		if schema, err = loader.Schema(ctx); err != nil {
			return nil, err
		}
	}
	// add schema to context
	ctx = context.WithValue(ctx, xo.SchemaKey, schema)

	return ctx, nil
}

// Args contains command-line arguments.
type Args struct {
	// Verbose enables verbose output.
	Verbose bool
	// DbParams are database parameters.
	DbParams DbParams
	// TemplateParams are template parameters.
	TemplateParams TemplateParams
	// QueryParams are query parameters.
	QueryParams QueryParams
	// SchemaParams are schema parameters.
	SchemaParams SchemaParams
	// OutParams are out parameters.
	OutParams OutParams
}

// NewArgs creates the command args.
func NewArgs(ctx context.Context, args *Args) (context.Context, *Args, error) {
	// kingpin settings
	kingpin.UsageTemplate(kingpin.CompactUsageTemplate)

	args.DbParams.Flags = make(map[xo.ContextKey]interface{})
	args.TemplateParams.Flags = make(map[xo.ContextKey]interface{})

	// add loader flags
	for key, v := range args.DbParams.Flags {
		// deref the interface (should always be a pointer to a type)
		ctx = context.WithValue(ctx, key, reflect.ValueOf(v).Elem().Interface())
	}
	// add gen type
	//ctx = context.WithValue(ctx, templates.GenTypeKey, cmd)

	// enable verbose output for sql queries
	if args.Verbose {
		models.SetLogger(func(s string, v ...interface{}) {
			fmt.Printf("SQL:\n%s\nPARAMS:\n%v\n\n", s, v)
		})
	}
	return ctx, args, nil
}

// DbParams are database parameters.
type DbParams struct {
	// Schema is the name of the database schema.
	Schema string
	// DSN is the database string (ie, postgres://user:pass@host:5432/dbname?args=)
	DSN string
	// Flags are additional loader flags.
	Flags map[xo.ContextKey]interface{}
}

// TemplateParams are template parameters.
type TemplateParams struct {
	// Type is the name of the template.
	Type string
	// Suffix is the file extension suffix.
	Suffix string
	// Src is the src directory of the template.
	Src string
	// Flags are additional template flags.
	Flags map[xo.ContextKey]interface{}
}

// QueryParams are query parameters.
type QueryParams struct {
	// Query is the passed query.
	//
	// If not specified, then os.Stdin will be used.
	Query string
	// Type is the type name.
	Type string
	// TypeComment is the type comment.
	TypeComment string
	// Func is the func name.
	Func string
	// FuncComment is the func comment.
	FuncComment string
	// Trim enables triming whitespace.
	Trim bool
	// Strip enables stripping the '::<type> AS <name>' in queries.
	Strip bool
	// One toggles the generated code to expect only one result.
	One bool
	// Flat toggles the generated code to return all scanned values directly.
	Flat bool
	// Exec toggles the generated code to do a db exec.
	Exec bool
	// Interpolate enables interpolation.
	Interpolate bool
	// Delimiter is the delimiter for parameterized values.
	Delimiter string
	// Fields are the fields to scan the result to.
	Fields string
	// AllowNulls toggles results can contain null types.
	AllowNulls bool
}

// SchemaParams are schema parameters.
type SchemaParams struct {
	// FkMode is the foreign resolution mode.
	FkMode string
	// Include allows the user to specify which types should be included. Can
	// match multiple types via regex patterns.
	//
	// - When unspecified, all types are included.
	// - When specified, only types match will be included.
	// - When a type matches an exclude entry and an include entry,
	//   the exclude entry will take precedence.
	Include []glob.Glob
	// Exclude allows the user to specify which types should be skipped. Can
	// match multiple types via regex patterns.
	//
	// When unspecified, all types are included in the schema.
	Exclude []glob.Glob
	// UseIndexNames toggles using index names.
	//
	// This is not enabled by default, because index names are often generated
	// using database design software which often gives non-descriptive names
	// to indexes (for example, 'authors__b124214__u_idx' instead of the more
	// descriptive 'authors_title_idx').
	UseIndexNames bool
}

// OutParams are out parameters.
type OutParams struct {
	// Out is the out path.
	Out string
	// Append toggles to append to the existing types.
	Append bool
	// Single when true changes behavior so that output is to one file.
	Single string
	// Debug toggles direct writing of files to disk, skipping post processing.
	Debug bool
}

// Open opens a connection to the database, returning a context for use in the
// application logic.
func Open(ctx context.Context, dsn, schema string) (context.Context, error) {
	v, err := user.Current()
	if err != nil {
		return nil, err
	}
	// parse dsn
	u, err := dburl.Parse(dsn)
	if err != nil {
		return nil, err
	}
	// open database
	db, err := passfile.OpenURL(u, v.HomeDir, "xopass")
	if err != nil {
		return nil, err
	}

	// add driver to context
	ctx = context.WithValue(ctx, xo.DriverKey, u.Driver)

	// add db to context
	ctx = context.WithValue(ctx, xo.DbKey, db)
	// determine schema
	if schema == "" {
		if schema, err = loader.Schema(ctx); err != nil {
			return nil, err
		}
	}
	// add schema to context
	ctx = context.WithValue(ctx, xo.SchemaKey, schema)
	return ctx, nil
}
