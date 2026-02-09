package mongodb

import (
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func BuildDatabaseOptions(
	opts *options.DatabaseOptions,
) options.Lister[options.DatabaseOptions] {
	dbOpts := options.Database()
	if opts.BSONOptions != nil {
		dbOpts = dbOpts.SetBSONOptions(opts.BSONOptions)
	}
	if opts.ReadConcern != nil {
		dbOpts = dbOpts.SetReadConcern(opts.ReadConcern)
	}
	if opts.ReadPreference != nil {
		dbOpts = dbOpts.SetReadPreference(opts.ReadPreference)
	}
	if opts.Registry != nil {
		dbOpts = dbOpts.SetRegistry(opts.Registry)
	}
	if opts.WriteConcern != nil {
		dbOpts = dbOpts.SetWriteConcern(opts.WriteConcern)
	}
	return dbOpts
}
