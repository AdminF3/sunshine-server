// Package graphql provides and implements a GraphQL schema with gqlgen.
//
// Code generation is instrumented by gqlgen.yml and the schema itself is
// defined in schema.graphql. After changing any of these files `go generate`
// is supposed to get invoked.
package graphql

//go:generate go run github.com/99designs/gqlgen
