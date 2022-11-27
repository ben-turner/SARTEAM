package models

type AssetState uint8

const (
	// AssetStateUnknown is the default state of an asset.
	AssetStateUnknown AssetState = iota
	// AssetStatePreparing is the state of an asset when it is preparing to be deployed.
	AssetStatePreparing
	// AssetStateReady indicates that an asset is ready and waiting to be tasked.
	AssetStateReady
	// AssetStateTasked indicates that an asset has been tasked and is currently in the field.
	AssetStateActive
	// AssetStateComplete indicates that an asset has completed its task.
	AssetStateComplete
)
