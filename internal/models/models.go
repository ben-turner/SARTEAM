// Package models contains the data structures used by the application. It also
// includes logic for manipulating the state, reading files and writing files to
// disk.
//
// State is stored on disk as a ledger of mutations. This so that incidents can
// be easily recovered in the event of an error, and for auditing purposes.
package models
