//go:build ignore
// +build ignore

package main

import (
	"log"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
)

func main() {
	// Generate Ent code with features
	if err := entc.Generate("./schema", &gen.Config{
		Features: []gen.Feature{
			gen.FeatureVersionedMigration, // Enables Atlas integration
			gen.FeaturePrivacy,
			gen.FeatureUpsert,
		},
	}); err != nil {
		log.Fatalf("running ent codegen: %v", err)
	}
}
