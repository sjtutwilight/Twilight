module github.com/sjtutwilight/Twilight/processor

go 1.22.5

require (
	github.com/ethereum/go-ethereum v1.13.5
	github.com/jmoiron/sqlx v1.4.0
	github.com/lib/pq v1.10.9
	github.com/segmentio/kafka-go v0.4.47
	github.com/sjtutwilight/Twilight/common v0.0.0
	github.com/stretchr/testify v1.8.4
)

replace github.com/sjtutwilight/Twilight/common => ../common

require (
	github.com/bits-and-blooms/bitset v1.7.0 // indirect
	github.com/btcsuite/btcd/btcec/v2 v2.2.0 // indirect
	github.com/consensys/bavard v0.1.13 // indirect
	github.com/consensys/gnark-crypto v0.12.1 // indirect
	github.com/crate-crypto/go-kzg-4844 v0.7.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.0.1 // indirect
	github.com/ethereum/c-kzg-4844 v0.4.0 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/holiman/uint256 v1.2.3 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/mmcloughlin/addchain v0.4.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/supranational/blst v0.3.11 // indirect
	golang.org/x/crypto v0.32.0 // indirect
	golang.org/x/sync v0.10.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	rsc.io/tmplfunc v0.0.3 // indirect
)
