# evmgo

> *Block explorers are for people who use the mouse.*

A CLI for querying the EVM. No browser required. No mouse involved.

## Current Scope

- Raw JSON-RPC passthrough
- Block, transaction, and receipt reads
- Native balance, code, and storage reads
- ABI-backed `eth_call`
- Log queries with optional event-name lookup from an ABI
- Human-readable text output by default
- Structured JSON output with `--json`

## Install

Download a prebuilt archive from [GitHub Releases](https://github.com/itzfelixv/evmgo/releases), or build from source:

```bash
git clone https://github.com/itzfelixv/evmgo.git
cd evmgo
go build ./cmd/evmgo
```

Or run it directly:

```bash
go run ./cmd/evmgo --help
```

## Quick Start

Set an RPC endpoint once:

```bash
export EVMGO_RPC="https://mainnet.infura.io/v3/<key>"
```

Then use the CLI:

```bash
evmgo block latest
evmgo bal 0x1111111111111111111111111111111111111111
evmgo tx 0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
evmgo receipt 0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
```

You can also pass the endpoint explicitly on every command:

```bash
evmgo --rpc https://mainnet.infura.io/v3/<key> block finalized
```

## Global Options

- `--rpc <url>`: JSON-RPC endpoint URL
- `--json`: emit machine-readable JSON instead of text
- `EVMGO_RPC`: fallback RPC endpoint when `--rpc` is not set

If neither `--rpc` nor `EVMGO_RPC` is set, the CLI exits with a structured error.

## Command Reference

### `rpc`

Send a raw JSON-RPC request.

```bash
evmgo rpc eth_blockNumber
evmgo rpc eth_getBalance 0x1111111111111111111111111111111111111111 latest
evmgo rpc eth_getBlockByNumber 0x10 false --json
```

Notes:

- Extra arguments are parsed as JSON literals when possible.
- If a param is not valid standalone JSON, it is sent as a string.
- Numbers preserve precision through `json.Number`.

### `block` / `blk`

Fetch a block by tag or number.

```bash
evmgo block latest
evmgo blk 42
evmgo block 0x2a --json
```

Accepted selectors:

- Decimal block number like `42`
- Hex block number like `0x2a`
- Tags: `latest`, `earliest`, `pending`, `safe`, `finalized`

### `tx`

Fetch a transaction by hash.

```bash
evmgo tx 0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
```

Validation:

- Transaction hashes must be full 32-byte hex hashes.

### `receipt` / `rcpt`

Fetch a transaction receipt by hash.

```bash
evmgo receipt 0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
evmgo rcpt 0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa --json
```

Validation:

- Transaction hashes must be full 32-byte hex hashes.

### `balance` / `bal`

Read the native balance for an address.

```bash
evmgo balance 0x1111111111111111111111111111111111111111
evmgo bal 0x1111111111111111111111111111111111111111 --json
```

Notes:

- Always reads at `latest`.

### `code`

Read contract bytecode.

```bash
evmgo code 0x2222222222222222222222222222222222222222
evmgo code 0x2222222222222222222222222222222222222222 --block finalized
```

### `storage`

Read a storage slot from a contract.

```bash
evmgo storage 0x2222222222222222222222222222222222222222 0x0
evmgo storage 0x2222222222222222222222222222222222222222 0x01 --block 42 --json
```

Validation:

- Address must be a valid hex address
- Slot must be a hex value
- Slot accepts either a hex integer like `0x0`/`0x2a` or a hex byte string up to 32 bytes

### `call`

Execute a read-only ABI-backed `eth_call`.

```bash
evmgo call \
  --to 0x2222222222222222222222222222222222222222 \
  --abi ./testdata/abi/erc20.json \
  --method balanceOf \
  --args 0x1111111111111111111111111111111111111111
```

Optional:

- `--block <selector>` to call against a specific block
- `--method` accepts a bare ABI method name when it is unique; for overloaded methods, pass the full signature such as `swap(uint256)`

`--args` accepts scalar ABI values directly. For composite values, pass a JSON literal in the matching ABI shape.

Example:

```bash
evmgo call \
  --to 0x2222222222222222222222222222222222222222 \
  --abi ./contract.json \
  --method setPair \
  --args '["0x1111111111111111111111111111111111111111",42]'
```

Current input coverage verified by tests includes:

- `address`
- `string`
- `bytes`
- `uint256[]`
- tuple inputs
- tuple-array inputs
- a dynamic tuple example

Current output:

- Raw return data is included
- Decoded values are rendered as strings
- Integer outputs are normalized to base-10 strings
- Composite outputs are rendered as JSON-style arrays in ABI field order

### `logs`

Query logs for a contract address and block range.

```bash
evmgo logs \
  --address 0x2222222222222222222222222222222222222222 \
  --from-block 18000000 \
  --to-block latest
```

With an ABI event name:

```bash
evmgo logs \
  --address 0x2222222222222222222222222222222222222222 \
  --abi ./testdata/abi/erc20.json \
  --event Transfer \
  --from-block 18000000 \
  --to-block latest
```

With additional topic filters:

```bash
evmgo logs \
  --address 0x2222222222222222222222222222222222222222 \
  --topic 0x0000000000000000000000001111111111111111111111111111111111111111 \
  --from-block 18000000 \
  --to-block 18001000
```

Notes:

- `--event` requires `--abi`
- `--event` accepts a bare ABI event name when it is unique; for overloaded events, pass the full signature such as `Filled(address)`
- `--event` derives `topic0` from the selected non-anonymous event signature
- Anonymous events are not supported with `--event`; use explicit `--topic` filters instead
- `--from-block` and `--to-block` default to `latest` when omitted
- In text mode, each log is printed as a single summary line

## Output Modes

Text is the default:

```text
address: 0x1111111111111111111111111111111111111111
balance: 0xde0b6b3a7640000
```

JSON is enabled with `--json`:

```bash
evmgo bal 0x1111111111111111111111111111111111111111 --json
```

Example:

```json
{
  "address": "0x1111111111111111111111111111111111111111",
  "balance": "0xde0b6b3a7640000"
}
```

Errors:

- Text mode writes `Error: ...` to stderr
- JSON mode writes a structured object like:

```json
{
  "error": "rpc endpoint required: pass --rpc or set EVMGO_RPC"
}
```

## ABI Files

For `call` and `logs --event`, you provide a standard ABI JSON file. Overloaded methods/events can be selected with full signatures. Anonymous events do not have signature-derived `topic0`, so filter them with `--topic` instead of `--event`.

This repo includes a small fixture at [testdata/abi/erc20.json](testdata/abi/erc20.json) with:

- `balanceOf(address) -> uint256`
- `Transfer(address,address,uint256)`

Example:

```bash
evmgo call \
  --to 0x2222222222222222222222222222222222222222 \
  --abi ./testdata/abi/erc20.json \
  --method balanceOf \
  --args 0x1111111111111111111111111111111111111111
```

## Development

Run the test suite:

```bash
go test ./...
```

Useful local checks:

```bash
go run ./cmd/evmgo --help
go run ./cmd/evmgo call --help
go run ./cmd/evmgo logs --help
```

ABI parsing, packing, event hashing, and decoding all live in the repo's internal ABI engine.

Current layout:

- `cmd/evmgo`: binary entrypoint
- `internal/cli`: Cobra commands and CLI validation
- `internal/config`: global flag resolution
- `internal/rpc`: JSON-RPC transport and block selectors
- `internal/actions`: command behavior
- `internal/abi`: ABI loading, packing, event signatures, and decoding
- `internal/output`: text and JSON rendering

## Contributing

Contributions are welcome. If you want to add or change behavior, open an issue or send a pull request with tests and a clear description of the change.

## License

`evmgo` is licensed under Apache-2.0. See [LICENSE](LICENSE).
