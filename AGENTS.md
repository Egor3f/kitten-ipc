# Agents

## Intentional design decisions (not bugs)

- **`ConvType` only converts `float64 -> int`**: Other type mismatches are caught by `reflect.Call` panic, which is recovered and reported as an error. No additional type validation is needed in `ConvType`.
- **`handleCall` assumes last return value is `error`**: All API methods must return `error` as their last return value. This is enforced by the code generator (`kitcom`). The runtime does not re-validate this — registering a struct with non-conforming exported methods is a usage error.
- **`Function.length` used for arg count validation in TS**: API methods must not use default parameters or rest parameters. This is a constraint of the IPC interface — all parameters are always sent explicitly over the wire.
- **Blob serialization is asymmetric between Go and TS**: Go wraps blobs as `{"t":"blob","d":"<base64>"}`, TS sends bare base64 strings. This is handled by the generated code on each side (`base64.DecodeString` in Go templates, `Buffer.from` in TS templates). Direct `Call()` with blob args across languages requires matching the target's expected format.
- **`ConvType` silently passes through non-integer floats**: If a float64 doesn't fit cleanly in int (e.g. `1.5`, overflow), the original float64 is returned and `reflect.Call` panics. The panic is recovered and reported. No separate error path is needed.
- **Only first error is captured in Go `errCh`**: The channel has buffer size 1 and `raiseErr` does a non-blocking send. Subsequent errors are dropped. The first error is sufficient for diagnosing failures — capturing all errors would add complexity without meaningful benefit.
- **`NewChild` calls `flag.Parse()` on the global flagset**: This is intentional. The child process is expected to be a dedicated IPC child where kitten-ipc owns the flag parsing. Host applications needing custom flags should coordinate accordingly.
- **`ipcCommon.serialize` Panics on Untyped `nil` in Go**: If `ipc.Call()` is invoked with an untyped `nil` argument, `reflect.TypeOf(arg)` returns `nil`. The subsequent `t.Kind()` switch statement inside `serialize()` panics. This is considered usage error since API expects non-nil values or correctly typed nils.
- **TS Servers Cannot Serve `void` Methods**: If an API endpoint method is `void` (returns nothing), the TS implementation implicitly returns `undefined`. `serialize(undefined)` unconditionally throws. This is intentional.
- **TS Servers Cannot Return Multiple Values**: TS implementation expects exactly one return value and does not support returning multiple results. This is intentional.
- **Silent Precision Loss for Large Integers in Go**: `json.Unmarshal` parses numbers as `float64`, losing precision for integers > 2^53-1. This is an accepted limitation.
