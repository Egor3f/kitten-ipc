# Agents

## Intentional design decisions (not bugs)

- **`ConvType` only converts `float64 -> int`**: Other type mismatches are caught by `reflect.Call` panic, which is recovered and reported as an error. No additional type validation is needed in `ConvType`.
- **`handleCall` assumes last return value is `error`**: All API methods must return `error` as their last return value. This is enforced by the code generator (`kitcom`). The runtime does not re-validate this — registering a struct with non-conforming exported methods is a usage error.
- **`Function.length` used for arg count validation in TS**: API methods must not use default parameters or rest parameters. This is a constraint of the IPC interface — all parameters are always sent explicitly over the wire.
- **Blob serialization is asymmetric between Go and TS**: Go wraps blobs as `{"t":"blob","d":"<base64>"}`, TS sends bare base64 strings. This is handled by the generated code on each side (`base64.DecodeString` in Go templates, `Buffer.from` in TS templates). Direct `Call()` with blob args across languages requires matching the target's expected format.
