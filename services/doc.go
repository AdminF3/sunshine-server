// Package services provides convenient access to external services.
//
// Currently managed services by this package:
//
// - Database store for each entity kind.
//
// - Sessions with secure cookie storage from gorilla.
//
// - Sending email with hermes.
//
// - Reporting errors to sentry.
//
// Access to all the services is being acquired via Env type which handles
// initialization of the underlying services without storing anything
// whatsoever in a global package scope.
package services
